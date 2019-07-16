package aws

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_s3_bucket_object", &resource.Sweeper{
		Name: "aws_s3_bucket_object",
		F:    testSweepS3BucketObjects,
	})
}

func testSweepS3BucketObjects(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).s3conn
	input := &s3.ListBucketsInput{}

	output, err := conn.ListBuckets(input)

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Bucket Objects sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing S3 Bucket Objects: %s", err)
	}

	if len(output.Buckets) == 0 {
		log.Print("[DEBUG] No S3 Bucket Objects to sweep")
		return nil
	}

	for _, bucket := range output.Buckets {
		bucketName := aws.StringValue(bucket.Name)

		hasPrefix := false
		prefixes := []string{"mybucket.", "mylogs.", "tf-acc", "tf-object-test", "tf-test"}

		for _, prefix := range prefixes {
			if strings.HasPrefix(bucketName, prefix) {
				hasPrefix = true
				break
			}
		}

		if !hasPrefix {
			log.Printf("[INFO] Skipping S3 Bucket: %s", bucketName)
			continue
		}

		input := &s3.ListObjectVersionsInput{
			Bucket: bucket.Name,
		}

		err := conn.ListObjectVersionsPages(input, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, objectVersion := range page.Versions {
				input := &s3.DeleteObjectInput{
					Bucket:    bucket.Name,
					Key:       objectVersion.Key,
					VersionId: objectVersion.VersionId,
				}
				objectKey := aws.StringValue(objectVersion.Key)
				objectVersionID := aws.StringValue(objectVersion.VersionId)

				log.Printf("[INFO] Deleting S3 Bucket (%s) Object (%s) Version: %s", bucketName, objectKey, objectVersionID)
				_, err := conn.DeleteObject(input)

				if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, s3.ErrCodeNoSuchKey, "") {
					continue
				}

				if err != nil {
					log.Printf("[ERROR] Error deleting S3 Bucket (%s) Object (%s) Version (%s): %s", bucketName, objectKey, objectVersionID, err)
				}
			}

			return !lastPage
		})

		if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
			continue
		}

		if isAWSErr(err, "AuthorizationHeaderMalformed", "region") || isAWSErr(err, "BucketRegionError", "") {
			log.Printf("[INFO] Skipping S3 Bucket (%s) Objects: %s", bucketName, err)
			continue
		}

		if err != nil {
			return fmt.Errorf("error listing S3 Bucket (%s) Objects: %s", bucketName, err)
		}

		err = conn.ListObjectVersionsPages(input, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, deleteMarker := range page.DeleteMarkers {
				input := &s3.DeleteObjectInput{
					Bucket:    bucket.Name,
					Key:       deleteMarker.Key,
					VersionId: deleteMarker.VersionId,
				}
				deleteMarkerKey := aws.StringValue(deleteMarker.Key)
				deleteMarkerVersionID := aws.StringValue(deleteMarker.VersionId)

				log.Printf("[INFO] Deleting S3 Bucket (%s) Object (%s) Delete Marker: %s", bucketName, deleteMarkerKey, deleteMarkerVersionID)
				_, err := conn.DeleteObject(input)

				if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, s3.ErrCodeNoSuchKey, "") {
					continue
				}

				if err != nil {
					log.Printf("[ERROR] Error deleting S3 Bucket (%s) Object (%s) Version (%s): %s", bucketName, deleteMarkerKey, deleteMarkerVersionID, err)
				}
			}

			return !lastPage
		})

		if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error listing S3 Bucket (%s) Objects: %s", bucketName, err)
		}
	}

	return nil
}

func TestAccAWSS3BucketObject_source(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	source := testAccAWSS3BucketObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfigSource(rInt, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, "{anything will do }"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_content(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfigContent(rInt, "some_bucket_content"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, "some_bucket_content"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_contentBase64(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfigContentBase64(rInt, base64.StdEncoding.EncodeToString([]byte("some_bucket_content"))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, "some_bucket_content"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_withContentCharacteristics(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	source := testAccAWSS3BucketObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_withContentCharacteristics(rInt, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectBody(&obj, "{anything will do }"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "binary/octet-stream"),
					resource.TestCheckResourceAttr(resourceName, "website_redirect", "http://google.com"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_updates(t *testing.T) {
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	sourceInitial := testAccAWSS3BucketObjectCreateTempFile(t, "initial object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccAWSS3BucketObjectCreateTempFile(t, "modified object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_updateable(rInt, false, sourceInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &originalObj),
					testAccCheckAWSS3BucketObjectBody(&originalObj, "initial object state"),
					resource.TestCheckResourceAttr(resourceName, "etag", "647d1d58e1011c743ec67d5e8af87b53"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_updateable(rInt, false, sourceModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &modifiedObj),
					testAccCheckAWSS3BucketObjectBody(&modifiedObj, "modified object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "1c7fd13df1515c2a13ad9eb068931f09"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_updatesWithVersioning(t *testing.T) {
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	sourceInitial := testAccAWSS3BucketObjectCreateTempFile(t, "initial versioned object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccAWSS3BucketObjectCreateTempFile(t, "modified versioned object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_updateable(rInt, true, sourceInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &originalObj),
					testAccCheckAWSS3BucketObjectBody(&originalObj, "initial versioned object state"),
					resource.TestCheckResourceAttr(resourceName, "etag", "cee4407fa91906284e2a5e5e03e86b1b"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_updateable(rInt, true, sourceModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &modifiedObj),
					testAccCheckAWSS3BucketObjectBody(&modifiedObj, "modified versioned object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "00b8c73b1b50e7cc932362c7225b8e29"),
					testAccCheckAWSS3BucketObjectVersionIdDiffers(&modifiedObj, &originalObj),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_kms(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	source := testAccAWSS3BucketObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withKMSId(rInt, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectSSE(resourceName, "aws:kms"),
					testAccCheckAWSS3BucketObjectBody(&obj, "{anything will do }"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_sse(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	source := testAccAWSS3BucketObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withSSE(rInt, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					testAccCheckAWSS3BucketObjectSSE(resourceName, "AES256"),
					testAccCheckAWSS3BucketObjectBody(&obj, "{anything will do }"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_acl(t *testing.T) {
	var obj1, obj2, obj3 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketObjectConfig_acl(rInt, "some_bucket_content", "private"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj1, "some_bucket_content"),
					resource.TestCheckResourceAttr(resourceName, "acl", "private"),
					testAccCheckAWSS3BucketObjectAcl(resourceName, []string{"FULL_CONTROL"}),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_acl(rInt, "some_bucket_content", "public-read"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj2),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj2, "some_bucket_content"),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
					testAccCheckAWSS3BucketObjectAcl(resourceName, []string{"FULL_CONTROL", "READ"}),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_acl(rInt, "changed_some_bucket_content", "private"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj3),
					testAccCheckAWSS3BucketObjectVersionIdDiffers(&obj3, &obj2),
					testAccCheckAWSS3BucketObjectBody(&obj3, "changed_some_bucket_content"),
					resource.TestCheckResourceAttr(resourceName, "acl", "private"),
					testAccCheckAWSS3BucketObjectAcl(resourceName, []string{"FULL_CONTROL"}),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_storageClass(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfigContent(rInt, "some_bucket_content"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "STANDARD"),
					testAccCheckAWSS3BucketObjectStorageClass(resourceName, "STANDARD"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_storageClass(rInt, "REDUCED_REDUNDANCY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "REDUCED_REDUNDANCY"),
					testAccCheckAWSS3BucketObjectStorageClass(resourceName, "REDUCED_REDUNDANCY"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_storageClass(rInt, "GLACIER"),
				Check: resource.ComposeTestCheckFunc(
					// Can't GetObject on an object in Glacier without restoring it.
					resource.TestCheckResourceAttr(resourceName, "storage_class", "GLACIER"),
					testAccCheckAWSS3BucketObjectStorageClass(resourceName, "GLACIER"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_storageClass(rInt, "INTELLIGENT_TIERING"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "storage_class", "INTELLIGENT_TIERING"),
					testAccCheckAWSS3BucketObjectStorageClass(resourceName, "INTELLIGENT_TIERING"),
				),
			},
			{
				Config: testAccAWSS3BucketObjectConfig_storageClass(rInt, "DEEP_ARCHIVE"),
				Check: resource.ComposeTestCheckFunc(
					// 	Can't GetObject on an object in DEEP_ARCHIVE without restoring it.
					resource.TestCheckResourceAttr(resourceName, "storage_class", "DEEP_ARCHIVE"),
					testAccCheckAWSS3BucketObjectStorageClass(resourceName, "DEEP_ARCHIVE"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_tags(t *testing.T) {
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()
	key := "test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withUpdatedTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj2),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "XXX"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "EEE"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withNoTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj3),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj3, &obj2),
					testAccCheckAWSS3BucketObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withTags(rInt, key, "changed stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj4),
					testAccCheckAWSS3BucketObjectVersionIdDiffers(&obj4, &obj3),
					testAccCheckAWSS3BucketObjectBody(&obj4, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
		},
	})
}

func TestAccAWSS3BucketObject_tagsLeadingSlash(t *testing.T) {
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	resourceName := "aws_s3_bucket_object.object"
	rInt := acctest.RandInt()
	key := "/test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withUpdatedTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj2),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckAWSS3BucketObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "XXX"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "EEE"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withNoTags(rInt, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj3),
					testAccCheckAWSS3BucketObjectVersionIdEquals(&obj3, &obj2),
					testAccCheckAWSS3BucketObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccAWSS3BucketObjectConfig_withTags(rInt, key, "changed stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &obj4),
					testAccCheckAWSS3BucketObjectVersionIdDiffers(&obj4, &obj3),
					testAccCheckAWSS3BucketObjectBody(&obj4, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
		},
	})
}

func testAccCheckAWSS3BucketObjectVersionIdDiffers(first, second *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if first.VersionId == nil {
			return fmt.Errorf("Expected first object to have VersionId: %s", first)
		}
		if second.VersionId == nil {
			return fmt.Errorf("Expected second object to have VersionId: %s", second)
		}

		if *first.VersionId == *second.VersionId {
			return fmt.Errorf("Expected Version IDs to differ, but they are equal (%s)", *first.VersionId)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketObjectVersionIdEquals(first, second *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if first.VersionId == nil {
			return fmt.Errorf("Expected first object to have VersionId: %s", first)
		}
		if second.VersionId == nil {
			return fmt.Errorf("Expected second object to have VersionId: %s", second)
		}

		if *first.VersionId != *second.VersionId {
			return fmt.Errorf("Expected Version IDs to be equal, but they differ (%s, %s)", *first.VersionId, *second.VersionId)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketObjectDestroy(s *terraform.State) error {
	s3conn := testAccProvider.Meta().(*AWSClient).s3conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_object" {
			continue
		}

		_, err := s3conn.HeadObject(
			&s3.HeadObjectInput{
				Bucket:  aws.String(rs.Primary.Attributes["bucket"]),
				Key:     aws.String(rs.Primary.Attributes["key"]),
				IfMatch: aws.String(rs.Primary.Attributes["etag"]),
			})
		if err == nil {
			return fmt.Errorf("AWS S3 Object still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAWSS3BucketObjectExists(n string, obj *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket Object ID is set")
		}

		s3conn := testAccProvider.Meta().(*AWSClient).s3conn
		out, err := s3conn.GetObject(
			&s3.GetObjectInput{
				Bucket:  aws.String(rs.Primary.Attributes["bucket"]),
				Key:     aws.String(rs.Primary.Attributes["key"]),
				IfMatch: aws.String(rs.Primary.Attributes["etag"]),
			})
		if err != nil {
			return fmt.Errorf("S3Bucket Object error: %s", err)
		}

		*obj = *out

		return nil
	}
}

func testAccCheckAWSS3BucketObjectBody(obj *s3.GetObjectOutput, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		body, err := ioutil.ReadAll(obj.Body)
		if err != nil {
			return fmt.Errorf("failed to read body: %s", err)
		}
		obj.Body.Close()

		if got := string(body); got != want {
			return fmt.Errorf("wrong result body %q; want %q", got, want)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketObjectAcl(n string, expectedPerms []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		s3conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := s3conn.GetObjectAcl(&s3.GetObjectAclInput{
			Bucket: aws.String(rs.Primary.Attributes["bucket"]),
			Key:    aws.String(rs.Primary.Attributes["key"]),
		})

		if err != nil {
			return fmt.Errorf("GetObjectAcl error: %v", err)
		}

		var perms []string
		for _, v := range out.Grants {
			perms = append(perms, *v.Permission)
		}
		sort.Strings(perms)

		if !reflect.DeepEqual(perms, expectedPerms) {
			return fmt.Errorf("Expected ACL permissions to be %v, got %v", expectedPerms, perms)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketObjectStorageClass(n, expectedClass string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		s3conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := s3conn.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(rs.Primary.Attributes["bucket"]),
			Key:    aws.String(rs.Primary.Attributes["key"]),
		})

		if err != nil {
			return fmt.Errorf("HeadObject error: %v", err)
		}

		// The "STANDARD" (which is also the default) storage
		// class when set would not be included in the results.
		storageClass := s3.StorageClassStandard
		if out.StorageClass != nil {
			storageClass = *out.StorageClass
		}

		if storageClass != expectedClass {
			return fmt.Errorf("Expected Storage Class to be %v, got %v",
				expectedClass, storageClass)
		}

		return nil
	}
}

func testAccCheckAWSS3BucketObjectSSE(n, expectedSSE string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		s3conn := testAccProvider.Meta().(*AWSClient).s3conn

		out, err := s3conn.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(rs.Primary.Attributes["bucket"]),
			Key:    aws.String(rs.Primary.Attributes["key"]),
		})

		if err != nil {
			return fmt.Errorf("HeadObject error: %v", err)
		}

		if out.ServerSideEncryption == nil {
			return fmt.Errorf("Expected a non %v Server Side Encryption.", out.ServerSideEncryption)
		}

		sse := *out.ServerSideEncryption
		if sse != expectedSSE {
			return fmt.Errorf("Expected Server Side Encryption %v, got %v.",
				expectedSSE, sse)
		}

		return nil
	}
}

func testAccAWSS3BucketObjectCreateTempFile(t *testing.T, data string) string {
	tmpFile, err := ioutil.TempFile("", "tf-acc-s3-obj")
	if err != nil {
		t.Fatal(err)
	}
	filename := tmpFile.Name()

	err = ioutil.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		os.Remove(filename)
		t.Fatal(err)
	}

	return filename
}

func testAccAWSS3BucketObjectConfigSource(randInt int, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket       = "${aws_s3_bucket.object_bucket.bucket}"
  key          = "test-key"
  source       = "%s"
  content_type = "binary/octet-stream"
}
`, randInt, source)
}

func testAccAWSS3BucketObjectConfig_withContentCharacteristics(randInt int, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket           = "${aws_s3_bucket.object_bucket.bucket}"
  key              = "test-key"
  source           = "%s"
  content_language = "en"
  content_type     = "binary/octet-stream"
  website_redirect = "http://google.com"
}
`, randInt, source)
}

func testAccAWSS3BucketObjectConfigContent(randInt int, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket  = "${aws_s3_bucket.object_bucket.bucket}"
  key     = "test-key"
  content = "%s"
}
`, randInt, content)
}

func testAccAWSS3BucketObjectConfigContentBase64(randInt int, contentBase64 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket         = "${aws_s3_bucket.object_bucket.bucket}"
  key            = "test-key"
  content_base64 = "%s"
}
`, randInt, contentBase64)
}

func testAccAWSS3BucketObjectConfig_updateable(randInt int, bucketVersioning bool, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket_3" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = %t
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket = "${aws_s3_bucket.object_bucket_3.bucket}"
  key    = "updateable-key"
  source = "%s"
  etag   = "${filemd5("%s")}"
}
`, randInt, bucketVersioning, source, source)
}

func testAccAWSS3BucketObjectConfig_withKMSId(randInt int, source string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "kms_key_1" {}

resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket     = "${aws_s3_bucket.object_bucket.bucket}"
  key        = "test-key"
  source     = "%s"
  kms_key_id = "${aws_kms_key.kms_key_1.arn}"
}
`, randInt, source)
}

func testAccAWSS3BucketObjectConfig_withSSE(randInt int, source string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket                 = "${aws_s3_bucket.object_bucket.bucket}"
  key                    = "test-key"
  source                 = "%s"
  server_side_encryption = "AES256"
}
`, randInt, source)
}

func testAccAWSS3BucketObjectConfig_acl(randInt int, content, acl string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket  = "${aws_s3_bucket.object_bucket.bucket}"
  key     = "test-key"
  content = "%s"
  acl     = "%s"
}
`, randInt, content, acl)
}

func testAccAWSS3BucketObjectConfig_storageClass(randInt int, storage_class string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object" {
  bucket        = "${aws_s3_bucket.object_bucket.bucket}"
  key           = "test-key"
  content       = "some_bucket_content"
  storage_class = "%s"
}
`, randInt, storage_class)
}

func testAccAWSS3BucketObjectConfig_withTags(randInt int, key, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket  = "${aws_s3_bucket.object_bucket.bucket}"
  key     = "%s"
  content = "%s"

  tags = {
    Key1 = "AAA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, randInt, key, content)
}

func testAccAWSS3BucketObjectConfig_withUpdatedTags(randInt int, key, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket  = "${aws_s3_bucket.object_bucket.bucket}"
  key     = "%s"
  content = "%s"

  tags = {
    Key2 = "BBB"
    Key3 = "XXX"
    Key4 = "DDD"
    Key5 = "EEE"
  }
}
`, randInt, key, content)
}

func testAccAWSS3BucketObjectConfig_withNoTags(randInt int, key, content string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket  = "${aws_s3_bucket.object_bucket.bucket}"
  key     = "%s"
  content = "%s"
}
`, randInt, key, content)
}
