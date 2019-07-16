package aws

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-template/template"
	"github.com/terraform-providers/terraform-provider-tls/tls"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvidersWithTLS map[string]terraform.ResourceProvider
var testAccProviderFactories func(providers *[]*schema.Provider) map[string]terraform.ResourceProviderFactory
var testAccProvider *schema.Provider
var testAccTemplateProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccTemplateProvider = template.Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"aws":      testAccProvider,
		"template": testAccTemplateProvider,
	}
	testAccProviderFactories = func(providers *[]*schema.Provider) map[string]terraform.ResourceProviderFactory {
		return map[string]terraform.ResourceProviderFactory{
			"aws": func() (terraform.ResourceProvider, error) {
				p := Provider()
				*providers = append(*providers, p.(*schema.Provider))
				return p, nil
			},
			"tls": func() (terraform.ResourceProvider, error) {
				p := tls.Provider()
				*providers = append(*providers, p.(*schema.Provider))
				return p, nil
			},
		}
	}
	testAccProvidersWithTLS = map[string]terraform.ResourceProvider{
		"tls": tls.Provider(),
	}

	for k, v := range testAccProviders {
		testAccProvidersWithTLS[k] = v
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("AWS_PROFILE") == "" && os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Fatal("AWS_ACCESS_KEY_ID or AWS_PROFILE must be set for acceptance tests")
	}

	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		t.Fatal("AWS_SECRET_ACCESS_KEY must be set for acceptance tests")
	}

	region := testAccGetRegion()
	log.Printf("[INFO] Test: Using %s as test region", region)
	os.Setenv("AWS_DEFAULT_REGION", region)

	err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	if err != nil {
		t.Fatal(err)
	}
}

// testAccAwsProviderAccountID returns the account ID of an AWS provider
func testAccAwsProviderAccountID(provider *schema.Provider) string {
	if provider == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: empty provider")
		return ""
	}
	if provider.Meta() == nil {
		log.Print("[DEBUG] Unable to read account ID from test provider: unconfigured provider")
		return ""
	}
	client, ok := provider.Meta().(*AWSClient)
	if !ok {
		log.Print("[DEBUG] Unable to read account ID from test provider: non-AWS or unconfigured AWS provider")
		return ""
	}
	return client.accountid
}

// testAccCheckResourceAttrAccountID ensures the Terraform state exactly matches the account ID
func testAccCheckResourceAttrAccountID(resourceName, attributeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestCheckResourceAttr(resourceName, attributeName, testAccGetAccountID())(s)
	}
}

// testAccCheckResourceAttrRegionalARN ensures the Terraform state exactly matches a formatted ARN with region
func testAccCheckResourceAttrRegionalARN(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: testAccGetAccountID(),
			Partition: testAccGetPartition(),
			Region:    testAccGetRegion(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// testAccMatchResourceAttrRegionalARN ensures the Terraform state regexp matches a formatted ARN with region
func testAccMatchResourceAttrRegionalARN(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: testAccGetAccountID(),
			Partition: testAccGetPartition(),
			Region:    testAccGetRegion(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %s", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// testAccCheckResourceAttrGlobalARN ensures the Terraform state exactly matches a formatted ARN without region
func testAccCheckResourceAttrGlobalARN(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: testAccGetAccountID(),
			Partition: testAccGetPartition(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// testAccMatchResourceAttrGlobalARN ensures the Terraform state regexp matches a formatted ARN without region
func testAccMatchResourceAttrGlobalARN(resourceName, attributeName, arnService string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			AccountID: testAccGetAccountID(),
			Partition: testAccGetPartition(),
			Resource:  arnResourceRegexp.String(),
			Service:   arnService,
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %s", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}

// testAccGetAccountID returns the account ID of testAccProvider
// Must be used returned within a resource.TestCheckFunc
func testAccGetAccountID() string {
	return testAccAwsProviderAccountID(testAccProvider)
}

func testAccGetRegion() string {
	v := os.Getenv("AWS_DEFAULT_REGION")
	if v == "" {
		return "us-west-2"
	}
	return v
}

func testAccGetPartition() string {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), testAccGetRegion()); ok {
		return partition.ID()
	}
	return "aws"
}

func testAccAlternateAccountPreCheck(t *testing.T) {
	if os.Getenv("AWS_ALTERNATE_PROFILE") == "" && os.Getenv("AWS_ALTERNATE_ACCESS_KEY_ID") == "" {
		t.Fatal("AWS_ALTERNATE_ACCESS_KEY_ID or AWS_ALTERNATE_PROFILE must be set for acceptance tests")
	}

	if os.Getenv("AWS_ALTERNATE_ACCESS_KEY_ID") != "" && os.Getenv("AWS_ALTERNATE_SECRET_ACCESS_KEY") == "" {
		t.Fatal("AWS_ALTERNATE_SECRET_ACCESS_KEY must be set for acceptance tests")
	}
}

func testAccEC2ClassicPreCheck(t *testing.T) {
	client := testAccProvider.Meta().(*AWSClient)
	platforms := client.supportedplatforms
	region := client.region
	if !hasEc2Classic(platforms) {
		t.Skipf("This test can only run in EC2 Classic, platforms available in %s: %q",
			region, platforms)
	}
}

func testAccHasServicePreCheck(service string, t *testing.T) {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), testAccGetRegion()); ok {
		if _, ok := partition.Services()[service]; !ok {
			t.Skip(fmt.Sprintf("skipping tests; partition does not support %s service", service))
		}
	}
}

func testAccMultipleRegionsPreCheck(t *testing.T) {
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), testAccGetRegion()); ok {
		if len(partition.Regions()) < 2 {
			t.Skip("skipping tests; partition only includes a single region")
		}
	}
}

func testAccOrganizationsAccountPreCheck(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn
	input := &organizations.DescribeOrganizationInput{}
	_, err := conn.DescribeOrganization(input)
	if isAWSErr(err, organizations.ErrCodeAWSOrganizationsNotInUseException, "") {
		return
	}
	if err != nil {
		t.Fatalf("error describing AWS Organization: %s", err)
	}
	t.Skip("skipping tests; this AWS account must not be an existing member of an AWS Organization")
}

func testAccAlternateAccountProviderConfig() string {
	return fmt.Sprintf(`
provider "aws" {
  access_key = %[1]q
  alias      = "alternate"
  profile    = %[2]q
  secret_key = %[3]q
}
`, os.Getenv("AWS_ALTERNATE_ACCESS_KEY_ID"), os.Getenv("AWS_ALTERNATE_PROFILE"), os.Getenv("AWS_ALTERNATE_SECRET_ACCESS_KEY"))
}

// Provider configuration hardcoded for us-east-1.
// This should only be necessary for testing ACM Certificates with CloudFront
// related infrastucture such as API Gateway Domain Names for EDGE endpoints,
// CloudFront Distribution Viewer Certificates, and Cognito User Pool Domains.
// Other valid usage is for services only available in us-east-1 such as the
// Cost and Usage Reporting and Pricing services.
func testAccUsEast1RegionProviderConfig() string {
	return fmt.Sprintf(`
provider "aws" {
  alias  = "us-east-1"
  region = "us-east-1"
}
`)
}

func testAccAwsRegionProviderFunc(region string, providers *[]*schema.Provider) func() *schema.Provider {
	return func() *schema.Provider {
		if region == "" {
			log.Println("[DEBUG] No region given")
			return nil
		}
		if providers == nil {
			log.Println("[DEBUG] No providers given")
			return nil
		}

		log.Printf("[DEBUG] Checking providers for AWS region: %s", region)
		for _, provider := range *providers {
			// Ignore if Meta is empty, this can happen for validation providers
			if provider == nil || provider.Meta() == nil {
				log.Printf("[DEBUG] Skipping empty provider")
				continue
			}

			// Ignore if Meta is not AWSClient, this will happen for other providers
			client, ok := provider.Meta().(*AWSClient)
			if !ok {
				log.Printf("[DEBUG] Skipping non-AWS provider")
				continue
			}

			clientRegion := client.region
			log.Printf("[DEBUG] Checking AWS provider region %q against %q", clientRegion, region)
			if clientRegion == region {
				log.Printf("[DEBUG] Found AWS provider with region: %s", region)
				return provider
			}
		}

		log.Printf("[DEBUG] No suitable provider found for %q in %d providers", region, len(*providers))
		return nil
	}
}

func testAccCheckWithProviders(f func(*terraform.State, *schema.Provider) error, providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		numberOfProviders := len(*providers)
		for i, provider := range *providers {
			if provider.Meta() == nil {
				log.Printf("[DEBUG] Skipping empty provider %d (total: %d)", i, numberOfProviders)
				continue
			}
			log.Printf("[DEBUG] Calling check with provider %d (total: %d)", i, numberOfProviders)
			if err := f(s, provider); err != nil {
				return err
			}
		}
		return nil
	}
}

// Check service API call error for reasons to skip acceptance testing
// These include missing API endpoints and unsupported API calls
func testAccPreCheckSkipError(err error) bool {
	// GovCloud has endpoints that respond with (no message provided after the error code):
	// AccessDeniedException:
	// Ignore these API endpoints that exist but are not officially enabled
	if isAWSErr(err, "AccessDeniedException", "") {
		return true
	}
	// Ignore missing API endpoints
	if isAWSErr(err, "RequestError", "send request failed") {
		return true
	}
	// Ignore unsupported API calls
	if isAWSErr(err, "UnknownOperationException", "") {
		return true
	}
	if isAWSErr(err, "UnsupportedOperation", "") {
		return true
	}
	return false
}

// Check sweeper API call error for reasons to skip sweeping
// These include missing API endpoints and unsupported API calls
func testSweepSkipSweepError(err error) bool {
	// Ignore missing API endpoints
	if isAWSErr(err, "RequestError", "send request failed") {
		return true
	}
	// Ignore unsupported API calls
	if isAWSErr(err, "UnsupportedOperation", "") {
		return true
	}
	// Ignore more unsupported API calls
	// InvalidParameterValue: Use of cache security groups is not permitted in this API version for your account.
	if isAWSErr(err, "InvalidParameterValue", "not permitted in this API version for your account") {
		return true
	}
	// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
	if isAWSErr(err, "InvalidParameterValue", "Access Denied to API Version") {
		return true
	}
	// GovCloud has endpoints that respond with (no message provided):
	// AccessDeniedException:
	// Since acceptance test sweepers are best effort and this response is very common,
	// we allow bypassing this error globally instead of individual test sweeper fixes.
	if isAWSErr(err, "AccessDeniedException", "") {
		return true
	}
	// Example: BadRequestException: vpc link not supported for region us-gov-west-1
	if isAWSErr(err, "BadRequestException", "not supported") {
		return true
	}
	// Example: InvalidAction: The action DescribeTransitGatewayAttachments is not valid for this web service
	if isAWSErr(err, "InvalidAction", "is not valid") {
		return true
	}
	return false
}

func TestAccAWSProvider_Endpoints(t *testing.T) {
	var providers []*schema.Provider
	var endpoints strings.Builder

	// Initialize each endpoint configuration with matching name and value
	for _, endpointServiceName := range endpointServiceNames {
		// Skip deprecated endpoint configurations as they will override expected values
		if endpointServiceName == "kinesis_analytics" || endpointServiceName == "r53" {
			continue
		}

		endpoints.WriteString(fmt.Sprintf("%s = \"http://%s\"\n", endpointServiceName, endpointServiceName))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigEndpoints(endpoints.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderEndpoints(&providers),
				),
			},
		},
	})
}

func TestAccAWSProvider_Endpoints_Deprecated(t *testing.T) {
	var providers []*schema.Provider
	var endpointsDeprecated strings.Builder

	// Initialize each deprecated endpoint configuration with matching name and value
	for _, endpointServiceName := range endpointServiceNames {
		// Only configure deprecated endpoint configurations
		if endpointServiceName != "kinesis_analytics" && endpointServiceName != "r53" {
			continue
		}

		endpointsDeprecated.WriteString(fmt.Sprintf("%s = \"http://%s\"\n", endpointServiceName, endpointServiceName))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSProviderConfigEndpoints(endpointsDeprecated.String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSProviderEndpointsDeprecated(&providers),
				),
			},
		},
	})
}

func testAccCheckAWSProviderEndpoints(providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		// Match AWSClient struct field names to endpoint configuration names
		endpointFieldNameF := func(endpoint string) func(string) bool {
			return func(name string) bool {
				switch endpoint {
				case "applicationautoscaling":
					endpoint = "appautoscaling"
				case "budgets":
					endpoint = "budget"
				case "cloudformation":
					endpoint = "cf"
				case "cloudhsm":
					endpoint = "cloudhsmv2"
				case "cognitoidentity":
					endpoint = "cognito"
				case "configservice":
					endpoint = "config"
				case "cur":
					endpoint = "costandusagereport"
				case "directconnect":
					endpoint = "dx"
				case "lexmodels":
					endpoint = "lexmodel"
				case "route53":
					endpoint = "r53"
				case "sdb":
					endpoint = "simpledb"
				case "serverlessrepo":
					endpoint = "serverlessapplicationrepository"
				case "servicecatalog":
					endpoint = "sc"
				case "servicediscovery":
					endpoint = "sd"
				case "stepfunctions":
					endpoint = "sfn"
				}

				switch name {
				case endpoint, fmt.Sprintf("%sconn", endpoint), fmt.Sprintf("%sConn", endpoint):
					return true
				}

				return false
			}
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*AWSClient)

			for _, endpointServiceName := range endpointServiceNames {
				// Skip deprecated endpoint configurations as they will override expected values
				if endpointServiceName == "kinesis_analytics" || endpointServiceName == "r53" {
					continue
				}

				providerClientField := reflect.Indirect(reflect.ValueOf(providerClient)).FieldByNameFunc(endpointFieldNameF(endpointServiceName))

				if !providerClientField.IsValid() {
					return fmt.Errorf("unable to match AWSClient struct field name for endpoint name: %s", endpointServiceName)
				}

				actualEndpoint := reflect.Indirect(reflect.Indirect(providerClientField).FieldByName("Config").FieldByName("Endpoint")).String()
				expectedEndpoint := fmt.Sprintf("http://%s", endpointServiceName)

				if actualEndpoint != expectedEndpoint {
					return fmt.Errorf("expected endpoint (%s) value (%s), got: %s", endpointServiceName, expectedEndpoint, actualEndpoint)
				}
			}
		}

		return nil
	}
}

func testAccCheckAWSProviderEndpointsDeprecated(providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if providers == nil {
			return fmt.Errorf("no providers initialized")
		}

		// Match AWSClient struct field names to endpoint configuration names
		endpointFieldNameF := func(endpoint string) func(string) bool {
			return func(name string) bool {
				switch endpoint {
				case "kinesis_analytics":
					endpoint = "kinesisanalytics"
				}

				return name == fmt.Sprintf("%sconn", endpoint)
			}
		}

		for _, provider := range *providers {
			if provider == nil || provider.Meta() == nil || provider.Meta().(*AWSClient) == nil {
				continue
			}

			providerClient := provider.Meta().(*AWSClient)

			for _, endpointServiceName := range endpointServiceNames {
				// Only check deprecated endpoint configurations
				if endpointServiceName != "kinesis_analytics" && endpointServiceName != "r53" {
					continue
				}

				providerClientField := reflect.Indirect(reflect.ValueOf(providerClient)).FieldByNameFunc(endpointFieldNameF(endpointServiceName))

				if !providerClientField.IsValid() {
					return fmt.Errorf("unable to match AWSClient struct field name for endpoint name: %s", endpointServiceName)
				}

				actualEndpoint := reflect.Indirect(reflect.Indirect(providerClientField).FieldByName("Config").FieldByName("Endpoint")).String()
				expectedEndpoint := fmt.Sprintf("http://%s", endpointServiceName)

				if actualEndpoint != expectedEndpoint {
					return fmt.Errorf("expected endpoint (%s) value (%s), got: %s", endpointServiceName, expectedEndpoint, actualEndpoint)
				}
			}
		}

		return nil
	}
}

func testAccAWSProviderConfigEndpoints(endpoints string) string {
	return fmt.Sprintf(`
provider "aws" {
  skip_credentials_validation = true
  skip_get_ec2_platforms      = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    %[1]s
  }
}

# Required to initialize the provider
data "aws_arn" "test" {
  arn = "arn:aws:s3:::test"
}
`, endpoints)
}
