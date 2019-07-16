package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
)

func TestGetSupportedEC2Platforms(t *testing.T) {
	ec2Endpoints := []*awsbase.MockEndpoint{
		{
			Request: &awsbase.MockRequest{
				Method: "POST",
				Uri:    "/",
				Body:   "Action=DescribeAccountAttributes&AttributeName.1=supported-platforms&Version=2016-11-15",
			},
			Response: &awsbase.MockResponse{
				StatusCode:  200,
				Body:        test_ec2_describeAccountAttributes_response,
				ContentType: "text/xml",
			},
		},
	}
	closeFunc, sess, err := awsbase.GetMockedAwsApiSession("EC2", ec2Endpoints)
	if err != nil {
		t.Fatal(err)
	}
	defer closeFunc()
	conn := ec2.New(sess)

	platforms, err := GetSupportedEC2Platforms(conn)
	if err != nil {
		t.Fatalf("Expected no error, received: %s", err)
	}
	expectedPlatforms := []string{"VPC", "EC2"}
	if !reflect.DeepEqual(platforms, expectedPlatforms) {
		t.Fatalf("Received platforms: %q\nExpected: %q\n", platforms, expectedPlatforms)
	}
}

var test_ec2_describeAccountAttributes_response = `<DescribeAccountAttributesResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
  <requestId>7a62c49f-347e-4fc4-9331-6e8eEXAMPLE</requestId>
  <accountAttributeSet>
    <item>
      <attributeName>supported-platforms</attributeName>
      <attributeValueSet>
        <item>
          <attributeValue>VPC</attributeValue>
        </item>
        <item>
          <attributeValue>EC2</attributeValue>
        </item>
      </attributeValueSet>
    </item>
  </accountAttributeSet>
</DescribeAccountAttributesResponse>`
