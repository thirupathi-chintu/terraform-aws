behavior "pull_request_path_labeler" "service_labels" {
  label_map = {
    # label provider related changes
    "provider" = [
      "aws/auth_helpers.go",
      "aws/awserr.go",
      "aws/config.go",
      "aws/*_aws_arn*",
      "aws/*_aws_ip_ranges*",
      "aws/*_aws_partition*",
      "aws/*_aws_region*",
      "aws/provider.go",
      "aws/utils.go",
      "website/docs/index.html.markdown",
      "website/**/arn*",
      "website/**/ip_ranges*",
      "website/**/partition*",
      "website/**/region*"
    ]
    # label test related changes
    "tests" = [
      "**/*_test.go",
      ".gometalinter.json",
      ".travis.yml"
    ]
    # label services
    "service/acm" = [
      "**/*_acm_*",
      "**/acm_*",
      "aws/tagsACM*"
    ]
    "service/acmpca" = [
      "**/*_acmpca_*",
      "**/acmpca_*",
      "aws/tagsACMPCA*"
    ]
    "service/alexaforbusiness" = [
      "**/*_alexaforbusiness_*",
      "**/alexaforbusiness_*"
    ]
    "service/amplify" = [
      "**/*_amplify_*",
      "**/amplify_*"
    ]
    "service/apigateway" = [
      "**/*_api_gateway_[^v][^2][^_]*",
      "**/api_gateway_[^v][^2][^_]*",
      "aws/tags_apigateway[^v][^2]*"
    ]
    "service/apigatewayv2" = [
      "**/*_api_gateway_v2_*",
      "**/api_gateway_v2_*",
      "aws/tags_apigatewayv2*"
    ]
    "service/applicationautoscaling" = [
      "**/*_appautoscaling_*",
      "**/appautoscaling_*"
    ]
    # "service/applicationdiscoveryservice" = [
    # 	"**/*_applicationdiscoveryservice_*",
    # 	"**/applicationdiscoveryservice_*"
    # ]
    "service/applicationinsights" = [
      "**/*_applicationinsights_*",
      "**/applicationinsights_*"
    ]
    "service/appmesh" = [
      "**/*_appmesh_*",
      "**/appmesh_*"
    ]
    "service/appstream" = [
      "**/*_appstream_*",
      "**/appstream_*"
    ]
    "service/appsync" = [
      "**/*_appsync_*",
      "**/appsync_*"
    ]
    "service/athena" = [
      "service/athena",
      "**/athena_*"
    ]
    "service/autoscaling" = [
      "**/*_autoscaling_*",
      "**/autoscaling_*",
      "aws/*_aws_launch_configuration*",
      "website/**/launch_configuration*"
    ]
    "service/autoscalingplans" = [
      "**/*_autoscalingplans_*",
      "**/autoscalingplans_*"
    ]
    "service/batch" = [
      "**/*_batch_*",
      "**/batch_*"
    ]
    "service/budgets" = [
      "**/*_budgets_*",
      "**/budgets_*"
    ]
    "service/cloud9" = [
      "**/*_cloud9_*",
      "**/cloud9_*"
    ]
    "service/clouddirectory" = [
      "**/*_clouddirectory_*",
      "**/clouddirectory_*"
    ]
    "service/cloudformation" = [
      "**/*_cloudformation_*",
      "**/cloudformation_*"
    ]
    "service/cloudfront" = [
      "**/*_cloudfront_*",
      "**/cloudfront_*",
      "aws/tagsCloudFront*"
    ]
    "service/cloudhsmv2" = [
      "**/*_cloudhsm_v2_*",
      "**/cloudhsm_v2_*"
    ]
    "service/cloudsearch" = [
      "**/*_cloudsearch_*",
      "**/cloudsearch_*"
    ]
    "service/cloudtrail" = [
      "**/*_cloudtrail_*",
      "**/cloudtrail_*",
      "aws/tagsCloudtrail*"
    ]
    "service/cloudwatch" = [
      "**/*_cloudwatch_dashboard*",
      "**/*_cloudwatch_metic_alarm*",
      "**/cloudwatch_dashboard*",
      "**/cloudwatch_metric_alarm*"
    ]
    "service/cloudwatchevents" = [
      "**/*_cloudwatch_event_*",
      "**/cloudwatch_event_*"
    ]
    "service/cloudwatchlogs" = [
      "**/*_cloudwatch_log_*",
      "**/cloudwatch_log_*"
    ]
    "service/codebuild" = [
      "**/*_codebuild_*",
      "**/codebuild_*",
      "aws/tagsCodeBuild*"
    ]
    "service/codecommit" = [
      "**/*_codecommit_*",
      "**/codecommit_*"
    ]
    "service/codedeploy" = [
      "**/*_codedeploy_*",
      "**/codedeploy_*"
    ]
    "service/codepipeline" = [
      "**/*_codepipeline_*",
      "**/codepipeline_*"
    ]
    "service/codestar" = [
      "**/*_codestar_*",
      "**/codestar_*"
    ]
    "service/cognito" = [
      "**/*_cognito_*",
      "**/_cognito_*"
    ]
    "service/configservice" = [
      "aws/*_aws_config_*",
      "website/**/config_*"
    ]
    "service/databasemigrationservice" = [
      "**/*_dms_*",
      "**/dms_*",
      "aws/tags_dms*"
    ]
    "service/datapipeline" = [
      "**/*_datapipeline_*",
      "**/datapipeline_*",
    ]
    "service/datasync" = [
      "**/*_datasync_*",
      "**/datasync_*",
    ]
    "service/dax" = [
      "**/*_dax_*",
      "**/dax_*",
      "aws/tagsDAX*"
    ]
    "service/devicefarm" = [
      "**/*_devicefarm_*",
      "**/devicefarm_*"
    ]
    "service/directconnect" = [
      "**/*_dx_*",
      "**/dx_*",
      "aws/tagsDX*"
    ]
    "service/directoryservice" = [
      "**/*_directory_service_*",
      "**/directory_service_*",
      "aws/tagsDS*"
    ]
    "service/dlm" = [
      "**/*_dlm_*",
      "**/dlm_*"
    ]
    "service/dynamodb" = [
      "**/*_dynamodb_*",
      "**/dynamodb_*"
    ]
    # Special casing this one because the files aren't _ec2_
    "service/ec2" = [
      "**/*_ec2_*",
      "**/ec2_*",
      "aws/*_aws_ami*",
      "aws/*_aws_availability_zone*",
      "aws/*_aws_customer_gateway*",
      "aws/*_aws_default_network_acl*",
      "aws/*_aws_default_route_table*",
      "aws/*_aws_default_security_group*",
      "aws/*_aws_default_subnet*",
      "aws/*_aws_default_vpc*",
      "aws/*_aws_ebs_*",
      "aws/*_aws_egress_only_internet_gateway*",
      "aws/*_aws_eip*",
      "aws/*_aws_flow_log*",
      "aws/*_aws_instance*",
      "aws/*_aws_internet_gateway*",
      "aws/*_aws_key_pair*",
      "aws/*_aws_launch_template*",
      "aws/*_aws_main_route_table_association*",
      "aws/*_aws_nat_gateway*",
      "aws/*_aws_network_acl*",
      "aws/*_aws_network_interface*",
      "aws/*_aws_placement_group*",
      "aws/*_aws_route_table*",
      "aws/*_aws_route.*",
      "aws/*_aws_security_group*",
      "aws/*_aws_spot*",
      "aws/*_aws_subnet*",
      "aws/*_aws_vpc*",
      "aws/*_aws_vpn*",
      "website/**/availability_zone*",
      "website/**/customer_gateway*",
      "website/**/default_network_acl*",
      "website/**/default_route_table*",
      "website/**/default_security_group*",
      "website/**/default_subnet*",
      "website/**/default_vpc*",
      "website/**/ebs_*",
      "website/**/egress_only_internet_gateway*",
      "website/**/eip*",
      "website/**/flow_log*",
      "website/**/instance*",
      "website/**/internet_gateway*",
      "website/**/key_pair*",
      "website/**/launch_template*",
      "website/**/main_route_table_association*",
      "website/**/nat_gateway*",
      "website/**/network_acl*",
      "website/**/network_interface*",
      "website/**/placement_group*",
      "website/**/route_table*",
      "website/**/route.*",
      "website/**/security_group*",
      "website/**/spot_*",
      "website/**/subnet.*",
      "website/**/vpc*",
      "website/**/vpn*"
    ]
    "service/ecr" = [
      "**/*_ecr_*",
      "**/ecr_*"
    ]
    "service/ecs" = [
      "**/*_ecs_*",
      "**/ecs_*"
    ]
    "service/efs" = [
      "**/*_efs_*",
      "**/efs_*",
      "aws/tagsEFS*"
    ]
    "service/eks" = [
      "**/*_eks_*",
      "**/eks_*"
    ]
    "service/elastic-transcoder" = [
      "**/*_elastic_transcoder_*",
      "**/elastic_transcoder_*"
    ]
    "service/elasticache" = [
      "**/*_elasticache_*",
      "**/elasticache_*",
      "aws/tagsEC*"
    ]
    "service/elasticbeanstalk" = [
      "**/*_elastic_beanstalk_*",
      "**/elastic_beanstalk_*",
      "aws/tagsBeanstalk*"
    ]
    "service/elasticsearch" = [
      "**/*_elasticsearch_*",
      "**/elasticsearch_*",
      "**/*_elasticsearchservice*"
    ]
    "service/elb" = [
      "aws/*_aws_app_cookie_stickiness_policy*",
      "aws/*_aws_elb*",
      "aws/*_aws_lb_cookie_stickiness_policy*",
      "aws/*_aws_lb_ssl_negotiation_policy*",
      "aws/*_aws_proxy_protocol_policy*",
      "aws/tagsELB*",
      "website/**/app_cookie_stickiness_policy*",
      "website/**/elb*",
      "website/**/lb_cookie_stickiness_policy*",
      "website/**/lb_ssl_negotiation_policy*",
      "website/**/proxy_protocol_policy*"
    ]
    "service/elbv2" = [
      "aws/*_lb.*",
      "aws/*_lb_listener*",
      "aws/*_lb_target_group*",
      "website/**/lb.*",
      "website/**/lb_listener*",
      "website/**/lb_target_group*"
    ]
    "service/emr" = [
      "**/*_emr_*",
      "**/emr_*"
    ]
    "service/firehose" = [
      "**/*_firehose_*",
      "**/firehose_*"
    ]
    "service/fms" = [
      "**/*_fms_*",
      "**/fms_*"
    ]
    "service/fsx" = [
      "**/*_fsx_*",
      "**/fsx_*"
    ]
    "service/gamelift" = [
      "**/*_gamelift_*",
      "**/gamelift_*"
    ]
    "service/glacier" = [
      "**/*_glacier_*",
      "**/glacier_*"
    ]
    "service/globalaccelerator" = [
      "**/*_globalaccelerator_*",
      "**/globalaccelerator_*"
    ]
    "service/glue" = [
      "**/*_glue_*",
      "**/glue_*"
    ]
    "service/greengrass" = [
      "**/*_greengrass_*",
      "**/greengrass_*"
    ]
    "service/guardduty" = [
      "**/*_guardduty_*",
      "**/guardduty_*"
    ]
    "service/iam" = [
      "**/*_iam_*",
      "**/iam_*"
    ]
    "service/inspector" = [
      "**/*_inspector_*",
      "**/inspector_*",
      "aws/tagsInspector*"
    ]
    "service/iot" = [
      "**/*_iot_*",
      "**/iot_*"
    ]
    "service/kinesis" = [
      "aws/*_aws_kinesis_stream*",
      "aws/tags_kinesis*",
      "website/kinesis_stream*"
    ]
    "service/kinesisanalytics" = [
      "**/*_kinesisanalytics_*",
      "**/kinesisanalytics_*"
    ]
    "service/kms" = [
      "**/*_kms_*",
      "**/kms_*",
      "aws/tagsKMS*"
    ]
    "service/lambda" = [
      "**/*_lambda_*",
      "**/lambda_*",
      "aws/tagsLambda*"
    ]
    "service/lexmodelbuildingservice" = [
      "**/*_lex_*",
      "**/lex_*"
    ]
    "service/licensemanager" = [
      "**/*_licensemanager_*",
      "**/licensemanager_*"
    ]
    "service/lightsail" = [
      "**/*_lightsail_*",
      "**/lightsail_*"
    ]
    "service/machinelearning" = [
      "**/*_machinelearning_*",
      "**/machinelearning_*"
    ]
    "service/macie" = [
      "**/*_macie_*",
      "**/macie_*"
    ]
    "service/mediaconnect" = [
      "**/*_media_connect_*",
      "**/media_connect_*"
    ]
    "service/mediaconvert" = [
      "**/*_media_convert_*",
      "**/media_convert_*"
    ]
    "service/medialive" = [
      "**/*_media_live_*",
      "**/media_live_*"
    ]
    "service/mediapackage" = [
      "**/*_media_package_*",
      "**/media_package_*"
    ]
    "service/mediastore" = [
      "**/*_media_store_*",
      "**/media_store_*"
    ]
    "service/mediatailor" = [
      "**/*_media_tailor_*",
      "**/media_tailor_*",
    ]
    "service/mobile" = [
      "**/*_mobile_*",
      "**/mobile_*"
    ],
    "service/mq" = [
      "**/*_mq_*",
      "**/mq_*"
    ]
    "service/neptune" = [
      "**/*_neptune_*",
      "**/neptune_*",
      "aws/tagsNeptune*"
    ]
    "service/opsworks" = [
      "**/*_opsworks_*",
      "**/opsworks_*",
      "aws/tagsOpsworks*"
    ]
    "service/organizations" = [
      "**/*_organizations_*",
      "**/organizations_*"
    ]
    "service/pinpoint" = [
      "**/*_pinpoint_*",
      "**/pinpoint_*"
    ]
    "service/polly" = [
      "**/*_polly_*",
      "**/polly_*"
    ]
    "service/pricing" = [
      "**/*_pricing_*",
      "**/pricing_*"
    ]
    "service/ram" = [
      "**/*_ram_*",
      "**/ram_*"
    ]
    "service/rds" = [
      "aws/*_aws_db_*",
      "aws/*_aws_rds_*",
      "aws/tagsRDS*",
      "website/**/db_*",
      "website/**/rds_*"
    ]
    "service/redshift" = [
      "**/*_redshift_*",
      "**/redshift_*",
      "aws/tagsRedshift*"
    ]
    "service/resourcegroups" = [
      "**/*_resourcegroups_*",
      "**/resourcegroups_*"
    ]
    "service/route53" = [
      "**/*_route53_delegation_set*",
      "**/*_route53_health_check*",
      "**/*_route53_query_log*",
      "**/*_route53_record*",
      "**/*_route53_zone*",
      "**/route53_delegation_set*",
      "**/route53_health_check*",
      "**/route53_query_log*",
      "**/route53_record*",
      "**/route53_zone*",
      "aws/tags_route53*"
    ]
    "service/robomaker" = [
      "**/*_robomaker_*",
      "**/robomaker_*",
    ]
    "service/route53domains" = [
      "**/*_route53_domains_*",
      "**/route53_domains_*"
    ]
    "service/s3" = [
      "**/*_s3_bucket*",
      "**/s3_bucket*",
      "aws/*_aws_canonical_user_id*",
      "website/**/canonical_user_id*"
    ]
    "service/s3control" = [
      "**/*_s3_account_*",
      "**/s3_account_*"
    ]
    "service/sagemaker" = [
      "**/*_sagemaker_*",
      "**/sagemaker_*"
    ]
    "service/secretsmanager" = [
      "**/*_secretsmanager_*",
      "**/secretsmanager_*",
      "aws/tagsSecretsManager*"
    ]
    "service/securityhub" = [
      "**/*_securityhub_*",
      "**/securityhub_*"
    ]
    "service/servicecatalog" = [
      "**/*_servicecatalog_*",
      "**/servicecatalog_*"
    ]
    "service/servicediscovery" = [
      "**/*_service_discovery_*",
      "**/service_discovery_*"
    ]
    "service/servicequotas" = [
      "**/*_servicequotas_*",
      "**/servicequotas_*"
    ]
    "service/ses" = [
      "**/*_ses_*",
      "**/ses_*",
      "aws/tagsSSM*"
    ]
    "service/sfn" = [
      "**/*_sfn_*",
      "**/sfn_*"
    ]
    "service/shield" = [
      "**/*_shield_*",
      "**/shield_*",
    ],
    "service/simpledb" = [
      "**/*_simpledb_*",
      "**/simpledb_*"
    ]
    "service/snowball" = [
      "**/*_snowball_*",
      "**/snowball_*"
    ]
    "service/sns" = [
      "**/*_sns_*",
      "**/sns_*"
    ]
    "service/sqs" = [
      "**/*_sqs_*",
      "**/sqs_*"
    ]
    "service/ssm" = [
      "**/*_ssm_*",
      "**/ssm_*"
    ]
    "service/storagegateway" = [
      "**/*_storagegateway_*",
      "**/storagegateway_*"
    ]
    "service/sts" = [
      "aws/*_aws_caller_identity*",
      "website/**/caller_identity*"
    ]
    "service/swf" = [
      "**/*_swf_*",
      "**/swf_*"
    ]
    "service/transfer" = [
      "**/*_transfer_*",
      "**/transfer_*"
    ]
    "service/waf" = [
      "**/*_waf_*",
      "**/waf_*",
      "**/*_wafregional_*",
      "**/wafregional_*"
    ]
    "service/workdocs" = [
      "**/*_workdocs_*",
      "**/workdocs_*"
    ]
    "service/workmail" = [
      "**/*_workmail_*",
      "**/workmail_*"
    ]
    "service/workspaces" = [
      "**/*_workspaces_*",
      "**/workspaces_*"
    ]
    "service/xray" = [
      "**/*_xray_*",
      "**/xray_*"
    ]
  }
}
