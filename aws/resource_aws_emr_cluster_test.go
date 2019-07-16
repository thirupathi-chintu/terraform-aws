package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_emr_cluster", &resource.Sweeper{
		Name: "aws_emr_cluster",
		F:    testSweepEmrClusters,
	})
}

func testSweepEmrClusters(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).emrconn

	input := &emr.ListClustersInput{
		ClusterStates: []*string{
			aws.String(emr.ClusterStateBootstrapping),
			aws.String(emr.ClusterStateRunning),
			aws.String(emr.ClusterStateStarting),
			aws.String(emr.ClusterStateWaiting),
		},
	}
	err = conn.ListClustersPages(input, func(page *emr.ListClustersOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, cluster := range page.Clusters {
			describeClusterInput := &emr.DescribeClusterInput{
				ClusterId: cluster.Id,
			}
			terminateJobFlowsInput := &emr.TerminateJobFlowsInput{
				JobFlowIds: []*string{cluster.Id},
			}
			id := aws.StringValue(cluster.Id)

			log.Printf("[INFO] Deleting EMR Cluster: %s", id)
			_, err = conn.TerminateJobFlows(terminateJobFlowsInput)

			if err != nil {
				log.Printf("[ERROR] Error terminating EMR Cluster (%s): %s", id, err)
			}

			if err := conn.WaitUntilClusterTerminated(describeClusterInput); err != nil {
				log.Printf("[ERROR] Error waiting for EMR Cluster (%s) termination: %s", id, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EMR Cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving EMR Clusters: %s", err)
	}

	return nil
}

func TestAccAWSEMRCluster_basic(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "scale_down_behavior", "TERMINATE_AT_TASK_COMPLETION"),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "step.#", "0"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_additionalInfo(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigAdditionalInfo(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "scale_down_behavior", "TERMINATE_AT_TASK_COMPLETION"),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "step.#", "0"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps", "additional_info"},
			},
		},
	})
}

func TestAccAWSEMRCluster_configurationsJson(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigConfigurationsJson(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestMatchResourceAttr("aws_emr_cluster.tf-test-cluster", "configurations_json",
						regexp.MustCompile("{\"JAVA_HOME\":\"/usr/lib/jvm/java-1.8.0\".+")),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_AutoscalingPolicy(t *testing.T) {
	var cluster1, cluster2, cluster3 emr.Cluster
	autoscalingPolicy1 := `
{
  "Constraints": {
    "MinCapacity": 1,
    "MaxCapacity": 2
  },
  "Rules": [
    {
      "Name": "ScaleOutMemoryPercentage",
      "Description": "Scale out if YARNMemoryAvailablePercentage is less than 15",
      "Action": {
        "SimpleScalingPolicyConfiguration": {
          "AdjustmentType": "CHANGE_IN_CAPACITY",
          "ScalingAdjustment": 1,
          "CoolDown": 300
        }
      },
      "Trigger": {
        "CloudWatchAlarmDefinition": {
          "ComparisonOperator": "LESS_THAN",
          "EvaluationPeriods": 1,
          "MetricName": "YARNMemoryAvailablePercentage",
          "Namespace": "AWS/ElasticMapReduce",
          "Period": 300,
          "Statistic": "AVERAGE",
          "Threshold": 15.0,
          "Unit": "PERCENT"
        }
      }
    }
  ]
}
`
	autoscalingPolicy2 := `
{
  "Constraints": {
    "MinCapacity": 1,
    "MaxCapacity": 3
  },
  "Rules": [
    {
      "Name": "ScaleOutMemoryPercentage",
      "Description": "Scale out if YARNMemoryAvailablePercentage is less than 15",
      "Action": {
        "SimpleScalingPolicyConfiguration": {
          "AdjustmentType": "CHANGE_IN_CAPACITY",
          "ScalingAdjustment": 1,
          "CoolDown": 300
        }
      },
      "Trigger": {
        "CloudWatchAlarmDefinition": {
          "ComparisonOperator": "LESS_THAN",
          "EvaluationPeriods": 1,
          "MetricName": "YARNMemoryAvailablePercentage",
          "Namespace": "AWS/ElasticMapReduce",
          "Period": 300,
          "Statistic": "AVERAGE",
          "Threshold": 15.0,
          "Unit": "PERCENT"
        }
      }
    }
  ]
}
`
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupAutoscalingPolicy(rName, autoscalingPolicy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "core_instance_group.0.autoscaling_policy", regexp.MustCompile(`"MaxCapacity": ?2`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupAutoscalingPolicy(rName, autoscalingPolicy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "core_instance_group.0.autoscaling_policy", regexp.MustCompile(`"MaxCapacity": ?3`)),
				),
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupAutoscalingPolicyRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster3),
					testAccCheckAWSEmrClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.autoscaling_policy", ""),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_BidPrice(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupBidPrice(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.bid_price", "0.50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupBidPrice(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.bid_price", "0.51"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_InstanceCount(t *testing.T) {
	var cluster1, cluster2, cluster3 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", "1"),
				),
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster3),
					testAccCheckAWSEmrClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", "2"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_InstanceType(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_type", "m4.large"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceType(rName, "m4.xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_type", "m4.xlarge"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_Migration_CoreInstanceType(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_type", "m4.large"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_type", "m4.large"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_Migration_InstanceGroup(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigInstanceGroupCoreInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "instance_group.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_type", "m4.large"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_CoreInstanceGroup_Name(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupName(rName, "name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigCoreInstanceGroupName(rName, "name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.name", "name2"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_instance_group(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigInstanceGroups(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "instance_group.#", "2"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_instance_group_names(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigInstanceGroupsName(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "instance_group.#", "3"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_instance_group_update(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigInstanceGroups(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "instance_group.#", "2"),
				),
			},
			{
				Config: testAccAWSEmrClusterConfigInstanceGroupsUpdate(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "instance_group.#", "2"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_instance_group_EBSVolumeType_st1(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigInstanceGroups_st1(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "instance_group.#", "2"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_updateAutoScalingPolicy(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigInstanceGroups_st1(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
				),
			},
			{
				Config: testAccAWSEmrClusterConfigInstanceGroups_updateAutoScalingPolicy(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_Kerberos_ClusterDedicatedKdc(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	password := fmt.Sprintf("NeverKeepPasswordsInPlainText%d!", r)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_Kerberos_ClusterDedicatedKdc(r, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "kerberos_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "kerberos_attributes.0.kdc_admin_password", password),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "kerberos_attributes.0.realm", "EC2.INTERNAL"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps", "kerberos_attributes.0.kdc_admin_password"},
			},
		},
	})
}

func TestAccAWSEMRCluster_MasterInstanceGroup_BidPrice(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupBidPrice(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.bid_price", "0.50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupBidPrice(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.bid_price", "0.51"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_MasterInstanceGroup_InstanceType(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_type", "m4.large"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupInstanceType(rName, "m4.xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_type", "m4.xlarge"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_MasterInstanceGroup_Migration_InstanceGroup(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigInstanceGroupMasterInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "instance_group.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_type", "m4.large"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_MasterInstanceGroup_Migration_MasterInstanceType(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_type", "m4.large"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_type", "m4.large"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_MasterInstanceGroup_Name(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupName(rName, "name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigMasterInstanceGroupName(rName, "name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					testAccCheckAWSEmrClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.name", "name2"),
				),
			},
		},
	})
}

func TestAccAWSEMRCluster_security_config(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_SecurityConfiguration(r),
				Check:  testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_Step_Basic(t *testing.T) {
	var cluster emr.Cluster
	rInt := acctest.RandInt()
	resourceName := "aws_emr_cluster.tf-test-cluster"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_Step_Single(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "step.0.action_on_failure", "TERMINATE_CLUSTER"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.args.0", "state-pusher-script"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.main_class", ""),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.properties.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "step.0.name", "Setup Hadoop Debugging"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_Step_ConfigMode(t *testing.T) {
	var cluster1, cluster2, cluster3 emr.Cluster
	rInt := acctest.RandInt()
	resourceName := "aws_emr_cluster.tf-test-cluster"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_Step_Single(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "step.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_state", "configurations", "keep_job_flow_alive_when_no_steps"},
			},
			{
				Config: testAccAWSEmrClusterConfig_Step_NoBlocks(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "step.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_state", "configurations", "keep_job_flow_alive_when_no_steps"},
			},
			{
				Config: testAccAWSEmrClusterConfig_Step_Zeroed(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "step.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cluster_state", "configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_Step_Multiple(t *testing.T) {
	var cluster emr.Cluster
	rInt := acctest.RandInt()
	resourceName := "aws_emr_cluster.tf-test-cluster"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_Step_Multiple(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "step.0.action_on_failure", "TERMINATE_CLUSTER"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.args.0", "state-pusher-script"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.0.name", "Setup Hadoop Debugging"),
					resource.TestCheckResourceAttr(resourceName, "step.1.action_on_failure", "CONTINUE"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.0", "spark-example"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.1", "SparkPi"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.2", "10"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.1.name", "Spark Step"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_bootstrap_ordering(t *testing.T) {
	var cluster emr.Cluster
	resourceName := "aws_emr_cluster.test"
	rName := acctest.RandomWithPrefix("tf-emr-bootstrap")
	argsInts := []string{
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
		"10",
	}

	argsStrings := []string{
		"instance.isMaster=true",
		"echo running on master node",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_bootstrap(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists(resourceName, &cluster),
					testAccCheck_bootstrap_order(&cluster, argsInts, argsStrings),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_terminationProtected(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigTerminationPolicy(r, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "termination_protection", "false"),
				),
			},
			{
				Config: testAccAWSEmrClusterConfigTerminationPolicy(r, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "termination_protection", "true"),
				),
			},
			{
				ResourceName:      "aws_emr_cluster.tf-test-cluster",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"kerberos_attributes.0.kdc_admin_password",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				//Need to turn off termination_protection to allow the job to be deleted
				Config: testAccAWSEmrClusterConfigTerminationPolicy(r, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "termination_protection", "false"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_keepJob(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig_keepJob(r, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "keep_job_flow_alive_when_no_steps", "false"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_visibleToAllUsers(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "visible_to_all_users", "true"),
				),
			},
			{
				ResourceName:      "aws_emr_cluster.tf-test-cluster",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations",
					"kerberos_attributes.0.kdc_admin_password",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccAWSEmrClusterConfigVisibleToAllUsersUpdated(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "visible_to_all_users", "false"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_s3Logging(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	bucketName := fmt.Sprintf("s3n://tf-acc-test-%d/", r)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigS3Logging(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "log_uri", bucketName),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_tags(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "tags.%", "4"),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "tags.role", "rolename"),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "tags.dns_zone", "env_zone"),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "tags.env", "env"),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "tags.name", "name-env")),
			},
			{
				Config: testAccAWSEmrClusterConfigUpdatedTags(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "tags.%", "3"),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "tags.dns_zone", "new_zone"),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "tags.Env", "production"),
					resource.TestCheckResourceAttr(
						"aws_emr_cluster.tf-test-cluster", "tags.name", "name-env"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_root_volume_size(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfig(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "ebs_root_volume_size", "21"),
				),
			},
			{
				Config: testAccAWSEmrClusterConfigUpdatedRootVolumeSize(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttr("aws_emr_cluster.tf-test-cluster", "ebs_root_volume_size", "48"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func TestAccAWSEMRCluster_custom_ami_id(t *testing.T) {
	var cluster emr.Cluster
	r := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrClusterConfigCustomAmiID(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrClusterExists("aws_emr_cluster.tf-test-cluster", &cluster),
					resource.TestCheckResourceAttrSet("aws_emr_cluster.tf-test-cluster", "custom_ami_id"),
				),
			},
			{
				ResourceName:            "aws_emr_cluster.tf-test-cluster",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configurations", "keep_job_flow_alive_when_no_steps"},
			},
		},
	})
}

func testAccCheck_bootstrap_order(cluster *emr.Cluster, argsInts, argsStrings []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		emrconn := testAccProvider.Meta().(*AWSClient).emrconn
		req := emr.ListBootstrapActionsInput{
			ClusterId: cluster.Id,
		}

		resp, err := emrconn.ListBootstrapActions(&req)
		if err != nil {
			return fmt.Errorf("Error listing boostrap actions in test: %s", err)
		}

		// make sure we actually checked something
		var ran bool
		for _, ba := range resp.BootstrapActions {
			// assume name matches the config
			rArgs := aws.StringValueSlice(ba.Args)
			if *ba.Name == "test" {
				ran = true
				if !reflect.DeepEqual(argsInts, rArgs) {
					return fmt.Errorf("Error matching Bootstrap args:\n\texpected: %#v\n\tgot: %#v", argsInts, rArgs)
				}
			} else if *ba.Name == "runif" {
				ran = true
				if !reflect.DeepEqual(argsStrings, rArgs) {
					return fmt.Errorf("Error matching Bootstrap args:\n\texpected: %#v\n\tgot: %#v", argsStrings, rArgs)
				}
			}
		}

		if !ran {
			return fmt.Errorf("Expected to compare bootstrap actions, but no checks were ran")
		}

		return nil
	}
}

func testAccCheckAWSEmrDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).emrconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emr_cluster" {
			continue
		}

		params := &emr.DescribeClusterInput{
			ClusterId: aws.String(rs.Primary.ID),
		}

		describe, err := conn.DescribeCluster(params)

		if err == nil {
			if describe.Cluster != nil &&
				*describe.Cluster.Status.State == "WAITING" {
				return fmt.Errorf("EMR Cluster still exists")
			}
		}

		providerErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		log.Printf("[ERROR] %v", providerErr)
	}

	return nil
}

func testAccCheckAWSEmrClusterExists(n string, v *emr.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No cluster id set")
		}
		conn := testAccProvider.Meta().(*AWSClient).emrconn
		describe, err := conn.DescribeCluster(&emr.DescribeClusterInput{
			ClusterId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("EMR error: %v", err)
		}

		if describe.Cluster == nil || *describe.Cluster.Id != rs.Primary.ID {
			return fmt.Errorf("EMR cluster %q not found", rs.Primary.ID)
		}

		*v = *describe.Cluster

		if describe.Cluster.Status != nil {
			state := aws.StringValue(describe.Cluster.Status.State)
			if state != emr.ClusterStateRunning && state != emr.ClusterStateWaiting {
				return fmt.Errorf("EMR cluster %q is not RUNNING or WAITING, currently: %s", rs.Primary.ID, state)
			}
		}

		return nil
	}
}

func testAccCheckAWSEmrClusterNotRecreated(i, j *emr.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Id) != aws.StringValue(j.Id) {
			return fmt.Errorf("EMR Cluster recreated: %s -> %s", aws.StringValue(i.Id), aws.StringValue(j.Id))
		}

		return nil
	}
}

func testAccCheckAWSEmrClusterRecreated(i, j *emr.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Id) == aws.StringValue(j.Id) {
			return fmt.Errorf("EMR Cluster not recreated: %s", aws.StringValue(i.Id))
		}

		return nil
	}
}

func testAccAWSEmrClusterConfigBaseVpc() string {
	return fmt.Sprintf(`
data "aws_availability_zones" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"

  ingress {
    from_port = 0
    protocol  = "-1"
    self      = true
    to_port   = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }

  # EMR will modify ingress rules
  lifecycle {
    ignore_changes = ["ingress"]
  }
}

resource "aws_subnet" "test" {
  availability_zone = "${data.aws_availability_zones.current.names[0]}"
  cidr_block        = "10.0.0.0/24"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-emr-cluster"
  }
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.test.id}"
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = "${aws_route_table.test.id}"
  subnet_id      = "${aws_subnet.test.id}"
}
`)
}

func testAccAWSEmrClusterConfig_bootstrap(r string) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  name                 = "%s"
  release_label        = "emr-5.0.0"
  applications         = ["Hadoop", "Hive"]
  log_uri              = "s3n://terraform/testlog/"
  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1
  service_role         = "${aws_iam_role.iam_emr_default_role.arn}"
  depends_on           = ["aws_main_route_table_association.a"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  bootstrap_action {
    path = "s3://${aws_s3_bucket.tester.bucket}/testscript.sh"
    name = "test"

    args = ["1",
      "2",
      "3",
      "4",
      "5",
      "6",
      "7",
      "8",
      "9",
      "10",
    ]
  }
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "%s_profile"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role" "iam_emr_default_role" {
  name = "%s_default_role"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role" "iam_emr_profile_role" {
  name = "%s_profile_role"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "%s_emr"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "%s_profile"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-bootstrap"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-bootstrap"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

output "cluser_id" {
  value = "${aws_emr_cluster.test.id}"
}

resource "aws_s3_bucket" "tester" {
  bucket = "%s"
  acl    = "public-read"
}

resource "aws_s3_bucket_object" "testobject" {
  bucket = "${aws_s3_bucket.tester.bucket}"
  key    = "testscript.sh"

  #source = "testscript.sh"
  content = "${data.template_file.testscript.rendered}"
  acl     = "public-read"
}

data "template_file" "testscript" {
  template = <<POLICY
#!/bin/bash
echo $@
POLICY
}
`, r, r, r, r, r, r, r)
}

func testAccAWSEmrClusterConfig(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  scale_down_behavior = "TERMINATE_AT_TASK_COMPLETION"

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role         = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role     = "${aws_iam_role.emr-autoscaling-role.arn}"
  ebs_root_volume_size = 21
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r)
}

func testAccAWSEmrClusterConfigAdditionalInfo(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  additional_info = <<EOF
{
	"instanceAwsClientConfiguration": {
		"proxyPort": 8099,
		"proxyHost": "myproxy.example.com"
	}
}
EOF

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  scale_down_behavior = "TERMINATE_AT_TASK_COMPLETION"

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role         = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role     = "${aws_iam_role.emr-autoscaling-role.arn}"
  ebs_root_volume_size = 21
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r)
}

func testAccAWSEmrClusterConfigConfigurationsJson(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-4.6.0"
  applications  = ["Hadoop", "Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations_json = <<EOF
 [
    {
      "Classification": "hadoop-env",
      "Configurations": [
        {
          "Classification": "export",
          "Properties": {
            "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
          }
        }
      ],
      "Properties": {}
    },
    {
      "Classification": "spark-env",
      "Configurations": [
        {
          "Classification": "export",
          "Properties": {
            "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
          }
        }
      ],
      "Properties": {}
    }
 ]
EOF

  depends_on = ["aws_main_route_table_association.a"]

  service_role         = "${aws_iam_role.iam_emr_default_role.arn}"
  ebs_root_volume_size = 21
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}
`, r)
}

func testAccAWSEmrClusterConfig_Kerberos_ClusterDedicatedKdc(r int, password string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_emr_security_configuration" "foo" {
  configuration = <<EOF
{
  "AuthenticationConfiguration": {
    "KerberosConfiguration": {
      "Provider": "ClusterDedicatedKdc",
      "ClusterDedicatedKdcConfiguration": {
        "TicketLifetimeInHours": 24
      }
    }
  }
}
EOF
}

resource "aws_emr_cluster" "tf-test-cluster" {
  applications                      = ["Spark"]
  core_instance_count               = 1
  core_instance_type                = "c4.large"
  keep_job_flow_alive_when_no_steps = true
  master_instance_type              = "c4.large"
  name                              = "emr-test-%[1]d"
  release_label                     = "emr-5.12.0"
  security_configuration            = "${aws_emr_security_configuration.foo.name}"
  service_role                      = "EMR_DefaultRole"
  termination_protection            = false

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.main.0.id}"
  }

  kerberos_attributes {
    kdc_admin_password = "%[2]s"
    realm              = "EC2.INTERNAL"
  }

  depends_on = ["aws_main_route_table_association.a"]
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-kerberos-cluster-dedicated-kdc"
  }
}

resource "aws_subnet" "main" {
  availability_zone = "${element(data.aws_availability_zones.available.names, count.index)}"
  count             = 2
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = "${aws_vpc.main.id}"

  tags = {
    Name = "tf-acc-emr-cluster-kerberos-cluster-dedicated-kdc"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"

  tags = {
    Name = "terraform-testacc-emr-cluster-kerberos-cluster-dedicated-kdc"
  }
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  route_table_id = "${aws_route_table.r.id}"
  vpc_id         = "${aws_vpc.main.id}"
}
`, r, password)
}

func testAccAWSEmrClusterConfig_SecurityConfiguration(rInt int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-5.5.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1

  security_configuration = "${aws_emr_security_configuration.foo.name}"

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role     = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role = "${aws_iam_role.emr-autoscaling-role.arn}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-security-configuration"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-security-configuration"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}

resource "aws_emr_security_configuration" "foo" {
  configuration = <<EOF
{
  "EncryptionConfiguration": {
    "AtRestEncryptionConfiguration": {
      "S3EncryptionConfiguration": {
        "EncryptionMode": "SSE-S3"
      },
      "LocalDiskEncryptionConfiguration": {
        "EncryptionKeyProviderType": "AwsKms",
        "AwsKmsKey": "${aws_kms_key.foo.arn}"
      }
    },
    "EnableInTransitEncryption": false,
    "EnableAtRestEncryption": true
  }
}
EOF
}

resource "aws_kms_key" "foo" {
  description             = "Terraform acc test %[1]d"
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}
`, rInt)
}

const testAccAWSEmrClusterConfig_Step_DebugLoggingStep = `
  # Example from: https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-plan-debugging.html
  step {
    action_on_failure = "TERMINATE_CLUSTER"
    name              = "Setup Hadoop Debugging"

    hadoop_jar_step {
      jar  = "command-runner.jar"
      args = ["state-pusher-script"]
    }
  }
`

const testAccAWSEmrClusterConfig_Step_SparkStep = `
  # Example from: https://docs.aws.amazon.com/emr/latest/ReleaseGuide/emr-spark-submit-step.html
  step {
    action_on_failure = "CONTINUE"
    name              = "Spark Step"

    hadoop_jar_step {
      jar  = "command-runner.jar"
      args = ["spark-example", "SparkPi", "10"]
    }
  }
`

func testAccAWSEmrClusterConfig_Step_Multiple(rInt int) string {
	stepConfig := testAccAWSEmrClusterConfig_Step_DebugLoggingStep + testAccAWSEmrClusterConfig_Step_SparkStep
	return testAccAWSEmrClusterConfig_Step(rInt, stepConfig)
}

func testAccAWSEmrClusterConfig_Step_NoBlocks(rInt int) string {
	return testAccAWSEmrClusterConfig_Step(rInt, "")
}

func testAccAWSEmrClusterConfig_Step_Single(rInt int) string {
	return testAccAWSEmrClusterConfig_Step(rInt, testAccAWSEmrClusterConfig_Step_DebugLoggingStep)
}

func testAccAWSEmrClusterConfig_Step_Zeroed(rInt int) string {
	return testAccAWSEmrClusterConfig_Step(rInt, "step = []")
}

func testAccAWSEmrClusterConfig_Step(rInt int, stepConfig string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_emr_cluster" "tf-test-cluster" {
  applications                      = ["Spark"]
  core_instance_count               = 1
  core_instance_type                = "c4.large"
  keep_job_flow_alive_when_no_steps = true
  log_uri                           = "s3://${aws_s3_bucket.test.bucket}/"
  master_instance_type              = "c4.large"
  name                              = "emr-test-%[1]d"
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"
  termination_protection            = false

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.main.0.id}"
  }

  %[2]s

  depends_on = ["aws_main_route_table_association.a"]
}

resource "aws_s3_bucket" "test" {
  bucket        = "tf-acc-test-%[1]d"
  force_destroy = true
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-step"
  }
}

resource "aws_subnet" "main" {
  availability_zone = "${element(data.aws_availability_zones.available.names, count.index)}"
  count             = 2
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = "${aws_vpc.main.id}"

  tags = {
    Name = "terraform-testacc-emr-cluster-step"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"

  tags = {
    Name = "terraform-testacc-emr-cluster-step"
  }
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  route_table_id = "${aws_route_table.r.id}"
  vpc_id         = "${aws_vpc.main.id}"
}
`, rInt, stepConfig)
}

func testAccAWSEmrClusterConfigCoreInstanceGroupAutoscalingPolicy(rName, autoscalingPolicy string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = [
        "application-autoscaling.amazonaws.com",
        "elasticmapreduce.amazonaws.com",
      ]
      type        = "Service"
    }
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = "${data.aws_iam_policy_document.test.json}"
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = "${aws_iam_role.test.name}"
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  autoscaling_role                  = "${aws_iam_role.test.arn}"
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    autoscaling_policy = <<POLICY%[2]sPOLICY
    instance_type      = "m4.large"
  }

  depends_on = ["aws_iam_role_policy_attachment.test", "aws_route_table_association.test"]
}
`, rName, autoscalingPolicy)
}

func testAccAWSEmrClusterConfigCoreInstanceGroupAutoscalingPolicyRemoved(rName string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = [
        "application-autoscaling.amazonaws.com",
        "elasticmapreduce.amazonaws.com",
      ]
      type        = "Service"
    }
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = "${data.aws_iam_policy_document.test.json}"
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = "${aws_iam_role.test.name}"
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  autoscaling_role                  = "${aws_iam_role.test.arn}"
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = "m4.large"
  }

  depends_on = ["aws_iam_role_policy_attachment.test", "aws_route_table_association.test"]
}
`, rName)
}

func testAccAWSEmrClusterConfigCoreInstanceGroupBidPrice(rName, bidPrice string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    bid_price     = %[2]q
    instance_type = "m4.large"
  }

  depends_on = ["aws_route_table_association.test"]
}
`, rName, bidPrice)
}

func testAccAWSEmrClusterConfigCoreInstanceGroupInstanceCount(rName string, instanceCount int) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_count = %[2]d
    instance_type  = "m4.large"
  }

  depends_on = ["aws_route_table_association.test"]
}
`, rName, instanceCount)
}

func testAccAWSEmrClusterConfigCoreInstanceGroupInstanceType(rName, instanceType string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = %[2]q
  }

  depends_on = ["aws_route_table_association.test"]
}
`, rName, instanceType)
}

func testAccAWSEmrClusterConfigCoreInstanceGroupName(rName, instanceGroupName string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = "m4.large"
    name          = %[2]q
  }

  depends_on = ["aws_route_table_association.test"]
}
`, rName, instanceGroupName)
}

func testAccAWSEmrClusterConfigCoreInstanceType(rName, coreInstanceType string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  core_instance_count               = 1
  core_instance_type                = %[2]q
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  depends_on = ["aws_route_table_association.test"]
}
`, rName, coreInstanceType)
}

func testAccAWSEmrClusterConfigInstanceGroupCoreInstanceType(rName, coreInstanceType string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  instance_group {
    instance_count = 1
    instance_role  = "MASTER"
    instance_type  = "m4.large"
  }

  instance_group {
    instance_count = 1
    instance_role  = "CORE"
    instance_type  = %[2]q
  }

  depends_on = ["aws_route_table_association.test"]
}
`, rName, coreInstanceType)
}

func testAccAWSEmrClusterConfigInstanceGroupMasterInstanceType(rName, masterInstanceType string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  instance_group {
    instance_count = 1
    instance_role  = "MASTER"
    instance_type  = %[2]q
  }

  depends_on = ["aws_route_table_association.test"]
}
`, rName, masterInstanceType)
}

func testAccAWSEmrClusterConfigInstanceGroups(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  instance_group {
    instance_role  = "CORE"
    instance_type  = "c4.large"
    instance_count = "1"

    ebs_config {
      size                 = "40"
      type                 = "gp2"
      volumes_per_instance = 1
    }

    bid_price = "0.30"

    autoscaling_policy = <<EOT
{
  "Constraints": {
    "MinCapacity": 1,
    "MaxCapacity": 2
  },
  "Rules": [
    {
      "Name": "ScaleOutMemoryPercentage",
      "Description": "Scale out if YARNMemoryAvailablePercentage is less than 15",
      "Action": {
        "SimpleScalingPolicyConfiguration": {
          "AdjustmentType": "CHANGE_IN_CAPACITY",
          "ScalingAdjustment": 1,
          "CoolDown": 300
        }
      },
      "Trigger": {
        "CloudWatchAlarmDefinition": {
          "ComparisonOperator": "LESS_THAN",
          "EvaluationPeriods": 1,
          "MetricName": "YARNMemoryAvailablePercentage",
          "Namespace": "AWS/ElasticMapReduce",
          "Period": 300,
          "Statistic": "AVERAGE",
          "Threshold": 15.0,
          "Unit": "PERCENT"
        }
      }
    }
  ]
}
EOT
  }

  instance_group {
    instance_role  = "MASTER"
    instance_type  = "c4.large"
    instance_count = 1
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role     = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role = "${aws_iam_role.emr-autoscaling-role.arn}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-instance-groups"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-instance-groups"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r)
}

func testAccAWSEmrClusterConfigInstanceGroupsName(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  instance_group {
    instance_count = 1
    instance_role  = "MASTER"
    instance_type  = "m3.xlarge"
    name           = "MasterInstanceGroup"
  }

  instance_group {
    instance_count = 2
    instance_role  = "CORE"
    instance_type  = "r4.xlarge"
    name           = "CoreInstanceGroup"

    ebs_config {
      iops                 = 0
      size                 = 100
      type                 = "gp2"
      volumes_per_instance = 1
    }
  }

  instance_group {
    instance_count = 3
    instance_role  = "TASK"
    instance_type  = "r4.xlarge"
    name           = "TaskInstanceGroup"

    ebs_config {
      iops                 = 0
      size                 = 100
      type                 = "gp2"
      volumes_per_instance = 1
    }
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role     = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role = "${aws_iam_role.emr-autoscaling-role.arn}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-instance-groups"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-instance-groups"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r)
}

func testAccAWSEmrClusterConfigInstanceGroupsUpdate(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  instance_group {
    instance_role  = "CORE"
    instance_type  = "c4.large"
    instance_count = "1"

    ebs_config {
      size                 = "40"
      type                 = "gp2"
      volumes_per_instance = 1
    }

    bid_price = "0.30"

    autoscaling_policy = <<EOT
{
  "Constraints": {
    "MinCapacity": 1,
    "MaxCapacity": 2
  },
  "Rules": [
    {
      "Name": "ScaleOutMemoryPercentage",
      "Description": "Scale out if YARNMemoryAvailablePercentage is less than 15",
      "Action": {
        "SimpleScalingPolicyConfiguration": {
          "AdjustmentType": "CHANGE_IN_CAPACITY",
          "ScalingAdjustment": 1,
          "CoolDown": 300
        }
      },
      "Trigger": {
        "CloudWatchAlarmDefinition": {
          "ComparisonOperator": "LESS_THAN",
          "EvaluationPeriods": 1,
          "MetricName": "YARNMemoryAvailablePercentage",
          "Namespace": "AWS/ElasticMapReduce",
          "Period": 300,
          "Statistic": "AVERAGE",
          "Threshold": 15.0,
          "Unit": "PERCENT"
        }
      }
    }
  ]
}
EOT
  }

  instance_group {
    instance_role  = "MASTER"
    instance_type  = "c4.large"
    instance_count = 1
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role     = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role = "${aws_iam_role.emr-autoscaling-role.arn}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-instance-groups"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-instance-groups"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r)
}

func testAccAWSEmrClusterConfigInstanceGroups_st1(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  instance_group {
    instance_role  = "CORE"
    instance_type  = "c4.large"
    instance_count = "1"

    ebs_config {
      size                 = "500"
      type                 = "st1"
      volumes_per_instance = 1
    }

    bid_price = "0.30"

    autoscaling_policy = <<EOT
{
  "Constraints": {
    "MinCapacity": 1,
    "MaxCapacity": 2
  },
  "Rules": [
    {
      "Name": "ScaleOutMemoryPercentage",
      "Description": "Scale out if YARNMemoryAvailablePercentage is less than 15",
      "Action": {
        "SimpleScalingPolicyConfiguration": {
          "AdjustmentType": "CHANGE_IN_CAPACITY",
          "ScalingAdjustment": 1,
          "CoolDown": 300
        }
      },
      "Trigger": {
        "CloudWatchAlarmDefinition": {
          "ComparisonOperator": "LESS_THAN",
          "EvaluationPeriods": 1,
          "MetricName": "YARNMemoryAvailablePercentage",
          "Namespace": "AWS/ElasticMapReduce",
          "Period": 300,
          "Statistic": "AVERAGE",
          "Threshold": 15.0,
          "Unit": "PERCENT"
        }
      }
    }
  ]
}
EOT
  }

  instance_group {
    instance_role  = "MASTER"
    instance_type  = "c4.large"
    instance_count = 1
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  depends_on = ["aws_main_route_table_association.a"]

  service_role     = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role = "${aws_iam_role.emr-autoscaling-role.arn}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-instance-groups"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-instance-groups"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r)
}

func testAccAWSEmrClusterConfigInstanceGroups_updateAutoScalingPolicy(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  instance_group {
    instance_role  = "CORE"
    instance_type  = "c4.large"
    instance_count = "1"

    ebs_config {
      size                 = "500"
      type                 = "st1"
      volumes_per_instance = 1
    }

    bid_price = "0.30"

    autoscaling_policy = <<EOT
{
  "Constraints": {
    "MinCapacity": 1,
    "MaxCapacity": 3
  },
  "Rules": [
    {
      "Name": "ScaleOutMemoryPercentage",
      "Description": "Scale out if YARNMemoryAvailablePercentage is less than 15",
      "Action": {
        "SimpleScalingPolicyConfiguration": {
          "AdjustmentType": "CHANGE_IN_CAPACITY",
          "ScalingAdjustment": 1,
          "CoolDown": 300
        }
      },
      "Trigger": {
        "CloudWatchAlarmDefinition": {
          "ComparisonOperator": "LESS_THAN",
          "EvaluationPeriods": 1,
          "MetricName": "YARNMemoryAvailablePercentage",
          "Namespace": "AWS/ElasticMapReduce",
          "Period": 300,
          "Statistic": "AVERAGE",
          "Threshold": 15.0,
          "Unit": "PERCENT"
        }
      }
    }
  ]
}
EOT
  }

  instance_group {
    instance_role  = "MASTER"
    instance_type  = "c4.large"
    instance_count = 1
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  depends_on = ["aws_main_route_table_association.a"]

  service_role     = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role = "${aws_iam_role.emr-autoscaling-role.arn}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-instance-groups"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-instance-groups"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r)
}

func testAccAWSEmrClusterConfigMasterInstanceGroupBidPrice(rName, bidPrice string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  master_instance_group {
    bid_price     = %[2]q
    instance_type = "m4.large"
  }

  depends_on = ["aws_route_table_association.test"]
}
`, rName, bidPrice)
}

func testAccAWSEmrClusterConfigMasterInstanceGroupInstanceType(rName, instanceType string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  master_instance_group {
    instance_type = %[2]q
  }

  depends_on = ["aws_route_table_association.test"]
}
`, rName, instanceType)
}

func testAccAWSEmrClusterConfigMasterInstanceGroupName(rName, instanceGroupName string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  master_instance_group {
    instance_type = "m4.large"
    name          = %[2]q
  }

  depends_on = ["aws_route_table_association.test"]
}
`, rName, instanceGroupName)
}

func testAccAWSEmrClusterConfigMasterInstanceType(rName, masterInstanceType string) string {
	return testAccAWSEmrClusterConfigBaseVpc() + fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  master_instance_type              = %[2]q
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  depends_on = ["aws_route_table_association.test"]
}
`, rName, masterInstanceType)
}

func testAccAWSEmrClusterConfigTerminationPolicy(r int, term string) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = %[2]s

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role     = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role = "${aws_iam_role.emr-autoscaling-role.arn}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-termination-policy"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-termination-policy"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r, term)
}

func testAccAWSEmrClusterConfig_keepJob(r int, keepJob string) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = %s
  termination_protection            = false

  step {
    action_on_failure = "CONTINUE"
    name              = "Sleep Step"

    hadoop_jar_step {
      jar  = "command-runner.jar"
      args = ["/bin/sleep", "60"]
    }
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role     = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role = "${aws_iam_role.emr-autoscaling-role.arn}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-termination-policy"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-termination-policy"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r, keepJob)
}

func testAccAWSEmrClusterConfigVisibleToAllUsersUpdated(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  visible_to_all_users              = false

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role     = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role = "${aws_iam_role.emr-autoscaling-role.arn}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-visible-to-all-users"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-visible-to-all-users"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r, r, r, r, r, r, r, r)
}

func testAccAWSEmrClusterConfigUpdatedTags(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%[1]d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1

  tags = {
    dns_zone = "new_zone"
    Env      = "production"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role     = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role = "${aws_iam_role.emr-autoscaling-role.arn}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%[1]d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-updated-tags"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-updated-tags"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%[1]d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%[1]d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%[1]d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%[1]d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r)
}

func testAccAWSEmrClusterConfigUpdatedRootVolumeSize(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role         = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role     = "${aws_iam_role.emr-autoscaling-role.arn}"
  ebs_root_volume_size = 48
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-updated-root-volume-size"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-updated-root-volume-size"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, r, r, r, r, r, r, r, r)
}

func testAccAWSEmrClusterConfigS3Logging(rInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "tf-acc-test-%d"
  force_destroy = true
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = "terraform-testacc-emr-cluster-s3-logging"
  }
}

resource "aws_subnet" "test" {
  vpc_id     = "${aws_vpc.test.id}"
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = "tf-acc-emr-cluster-s3-logging"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.main.id}"
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = "${aws_subnet.test.id}"
  route_table_id = "${aws_route_table.test.id}"
}

resource "aws_security_group" "test" {
  name        = "tf-acc-test-%d"
  description = "tf acceptance test"
  vpc_id      = "${aws_vpc.test.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "tf-acc-test-%d"
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  termination_protection            = false
  keep_job_flow_alive_when_no_steps = true

  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1

  log_uri = "s3://${aws_s3_bucket.test.bucket}/"

  ec2_attributes {
    instance_profile                  = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:instance-profile/EMR_EC2_DefaultRole"
    emr_managed_master_security_group = "${aws_security_group.test.id}"
    emr_managed_slave_security_group  = "${aws_security_group.test.id}"
    subnet_id                         = "${aws_subnet.test.id}"
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  service_role = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/EMR_DefaultRole"
}

data "aws_caller_identity" "current" {}
`, rInt, rInt, rInt)
}

func testAccAWSEmrClusterConfigCustomAmiID(r int) string {
	return fmt.Sprintf(`
resource "aws_emr_cluster" "tf-test-cluster" {
  name          = "emr-test-%d"
  release_label = "emr-5.7.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = "${aws_subnet.main.id}"
    emr_managed_master_security_group = "${aws_security_group.allow_all.id}"
    emr_managed_slave_security_group  = "${aws_security_group.allow_all.id}"
    instance_profile                  = "${aws_iam_instance_profile.emr_profile.arn}"
  }

  master_instance_type = "c4.large"
  core_instance_type   = "c4.large"
  core_instance_count  = 1

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = ["aws_main_route_table_association.a"]

  service_role         = "${aws_iam_role.iam_emr_default_role.arn}"
  autoscaling_role     = "${aws_iam_role.emr-autoscaling-role.arn}"
  ebs_root_volume_size = 48
  custom_ami_id        = "${data.aws_ami.emr-custom-ami.id}"
}

resource "aws_security_group" "allow_all" {
  name        = "allow_all_%d"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = ["aws_subnet.main"]

  lifecycle {
    ignore_changes = ["ingress", "egress"]
  }

  tags = {
    Name = "emr_test"
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-emr-cluster-custom-ami-id"
  }
}

resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "168.31.0.0/20"

  tags = {
    Name = "tf-acc-emr-cluster-custom-ami-id"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.main.id}"
  route_table_id = "${aws_route_table.r.id}"
}

###

# IAM things

###

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_default_role" {
  name = "iam_emr_default_role_%d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "service-attach" {
  role       = "${aws_iam_role.iam_emr_default_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_default_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_default_policy" {
  name = "iam_emr_default_policy_%d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeImages",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "iam_emr_profile_role_%d"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "emr_profile_%d"
  role = "${aws_iam_role.iam_emr_profile_role.name}"
}

resource "aws_iam_role_policy_attachment" "profile-attach" {
  role       = "${aws_iam_role.iam_emr_profile_role.id}"
  policy_arn = "${aws_iam_policy.iam_emr_profile_policy.arn}"
}

resource "aws_iam_policy" "iam_emr_profile_policy" {
  name = "iam_emr_profile_policy_%d"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "emr-autoscaling-role" {
  name               = "EMR_AutoScaling_DefaultRole_%d"
  assume_role_policy = "${data.aws_iam_policy_document.emr-autoscaling-role-policy.json}"
}

data "aws_iam_policy_document" "emr-autoscaling-role-policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.amazonaws.com", "application-autoscaling.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr-autoscaling-role" {
  role       = "${aws_iam_role.emr-autoscaling-role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}

data "aws_ami" "emr-custom-ami" {
  most_recent = true
  owners      = ["137112412989"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}
`, r, r, r, r, r, r, r, r)
}
