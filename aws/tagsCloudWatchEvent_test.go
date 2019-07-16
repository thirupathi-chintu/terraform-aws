package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

func TestDiffCloudWatchEventTags(t *testing.T) {
	cases := []struct {
		Old, New       map[string]interface{}
		Create, Remove map[string]string
	}{
		// Basic add/remove
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"bar": "baz",
			},
			Create: map[string]string{
				"bar": "baz",
			},
			Remove: map[string]string{
				"foo": "bar",
			},
		},

		// Modify
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"foo": "baz",
			},
			Create: map[string]string{
				"foo": "baz",
			},
			Remove: map[string]string{
				"foo": "bar",
			},
		},
	}

	for i, tc := range cases {
		c, r := diffTagsCloudWatchEvents(tagsFromMapCloudWatchEvents(tc.Old), tagsFromMapCloudWatchEvents(tc.New))
		cm := tagsToMapCloudWatchEvents(c)
		rm := tagsToMapCloudWatchEvents(r)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
	}
}

func TestIgnoringTagsCloudWatchEvents(t *testing.T) {
	var ignoredTags []*events.Tag
	ignoredTags = append(ignoredTags, &events.Tag{
		Key:   aws.String("aws:cloudformation:logical-id"),
		Value: aws.String("foo"),
	})
	ignoredTags = append(ignoredTags, &events.Tag{
		Key:   aws.String("aws:foo:bar"),
		Value: aws.String("baz"),
	})
	for _, tag := range ignoredTags {
		if !tagIgnoredCloudWatchEvents(tag) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", *tag.Key, *tag.Value)
		}
	}
}
