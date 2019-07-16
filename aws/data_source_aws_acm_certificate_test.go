package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/resource"
)

const ACMCertificateRe = `^arn:[^:]+:acm:[^:]+:[^:]+:certificate/.+$`

func TestAccAWSAcmCertificateDataSource_singleIssued(t *testing.T) {
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	var arnRe *regexp.Regexp
	var domain string

	if os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_MOST_RECENT_ARN") != "" {
		arnRe = regexp.MustCompile(fmt.Sprintf("^%s$", os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_MOST_RECENT_ARN")))
	} else {
		arnRe = regexp.MustCompile(ACMCertificateRe)
	}

	if os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_DOMAIN") != "" {
		domain = os.Getenv("ACM_CERTIFICATE_SINGLE_ISSUED_DOMAIN")
	} else {
		domain = fmt.Sprintf("tf-acc-single-issued.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))
	}

	resourceName := "data.aws_acm_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAcmCertificateDataSourceConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCheckAwsAcmCertificateDataSourceConfigWithStatus(domain, acm.CertificateStatusIssued),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCheckAwsAcmCertificateDataSourceConfigWithTypes(domain, acm.CertificateTypeAmazonIssued),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecent(domain, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecentAndStatus(domain, acm.CertificateStatusIssued, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecentAndTypes(domain, acm.CertificateTypeAmazonIssued, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
		},
	})
}

func TestAccAWSAcmCertificateDataSource_multipleIssued(t *testing.T) {
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	var arnRe *regexp.Regexp
	var domain string

	if os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_MOST_RECENT_ARN") != "" {
		arnRe = regexp.MustCompile(fmt.Sprintf("^%s$", os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_MOST_RECENT_ARN")))
	} else {
		arnRe = regexp.MustCompile(ACMCertificateRe)
	}

	if os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_DOMAIN") != "" {
		domain = os.Getenv("ACM_CERTIFICATE_MULTIPLE_ISSUED_DOMAIN")
	} else {
		domain = fmt.Sprintf("tf-acc-multiple-issued.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))
	}

	resourceName := "data.aws_acm_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckAwsAcmCertificateDataSourceConfig(domain),
				ExpectError: regexp.MustCompile(`Multiple certificates for domain`),
			},
			{
				Config:      testAccCheckAwsAcmCertificateDataSourceConfigWithStatus(domain, acm.CertificateStatusIssued),
				ExpectError: regexp.MustCompile(`Multiple certificates for domain`),
			},
			{
				Config:      testAccCheckAwsAcmCertificateDataSourceConfigWithTypes(domain, acm.CertificateTypeAmazonIssued),
				ExpectError: regexp.MustCompile(`Multiple certificates for domain`),
			},
			{
				Config: testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecent(domain, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecentAndStatus(domain, acm.CertificateStatusIssued, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
			{
				Config: testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecentAndTypes(domain, acm.CertificateTypeAmazonIssued, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, "arn", arnRe),
				),
			},
		},
	})
}

func TestAccAWSAcmCertificateDataSource_noMatchReturnsError(t *testing.T) {
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	domain := fmt.Sprintf("tf-acc-nonexistent.%s", os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckAwsAcmCertificateDataSourceConfig(domain),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
			{
				Config:      testAccCheckAwsAcmCertificateDataSourceConfigWithStatus(domain, acm.CertificateStatusIssued),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
			{
				Config:      testAccCheckAwsAcmCertificateDataSourceConfigWithTypes(domain, acm.CertificateTypeAmazonIssued),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
			{
				Config:      testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecent(domain, true),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
			{
				Config:      testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecentAndStatus(domain, acm.CertificateStatusIssued, true),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
			{
				Config:      testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecentAndTypes(domain, acm.CertificateTypeAmazonIssued, true),
				ExpectError: regexp.MustCompile(`No certificate for domain`),
			},
		},
	})
}

func TestAccAWSAcmCertificateDataSource_KeyTypes(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	dataSourceName := "data.aws_acm_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProvidersWithTLS,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmCertificateDataSourceConfigKeyTypes(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
				),
			},
		},
	})
}

func testAccCheckAwsAcmCertificateDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain = "%s"
}
`, domain)
}

func testAccCheckAwsAcmCertificateDataSourceConfigWithStatus(domain, status string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain   = "%s"
  statuses = ["%s"]
}
`, domain, status)
}

func testAccCheckAwsAcmCertificateDataSourceConfigWithTypes(domain, certType string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain = "%s"
  types  = ["%s"]
}
`, domain, certType)
}

func testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecent(domain string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
	domain = "%s"
	most_recent = %v
}
`, domain, mostRecent)
}

func testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecentAndStatus(domain, status string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
	domain = "%s"
	statuses = ["%s"]
	most_recent = %v
}
`, domain, status, mostRecent)
}

func testAccCheckAwsAcmCertificateDataSourceConfigWithMostRecentAndTypes(domain, certType string, mostRecent bool) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
	domain = "%s"
	types = ["%s"]
	most_recent = %v
}
`, domain, certType, mostRecent)
}

func testAccAwsAcmCertificateDataSourceConfigKeyTypes() string {
	return fmt.Sprintf(`
resource "tls_private_key" "test" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_self_signed_cert" "test" {
  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]

  key_algorithm         = "RSA"
  private_key_pem       = "${tls_private_key.test.private_key_pem}"
  validity_period_hours = 12

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }
}

resource "aws_acm_certificate" "test" {
  certificate_body = "${tls_self_signed_cert.test.cert_pem}"
  private_key      = "${tls_private_key.test.private_key_pem}"
}

data "aws_acm_certificate" "test" {
  domain    = "${aws_acm_certificate.test.domain_name}"
  key_types = ["RSA_4096"]
}
`)
}
