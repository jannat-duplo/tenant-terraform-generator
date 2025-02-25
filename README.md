# Terraform code generation for a DuploCloud Tenant

This utility provides a way to export the terraform code that represents the infrastructure deployed in a DuploCloud Tenant. This is often very useful in order to:
- Generate and persist DuploCloud Terraform IaC which can be version controlled in the future.
- Clone a new Tenant based on an already existing Tenant. 

## Prerequisite

1. Install [Go](https://go.dev/doc/install)
2. Install [make](https://www.gnu.org/software/make) tool.
3. Install [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli) version greater than or equals to `v0.14.11`
4. Following environment variables to be exported in the shell while running this projects.

```shell
# Required Vars
export customer_name="duplo-masp"
export tenant_name="test"
export cert_arn="arn:aws:acm:us-west-2:128329325849:certificate/1234567890-aaaa-bbbb-ccc-66e7dcd609e1"
export duplo_host="https://msp.duplocloud.net"
export duplo_token="xxx-xxxxx-xxxxxxxx"
export AWS_RUNNER="duplo-admin"
export aws_account_id="1234567890"
```

## How to run this project to export DuploCloud Provider terraform code?

- Clone this repository.

- Prepare environment variables and export within the shell as mentioned above.

- Run using  following command

  ```shell
  make run
  ```

- **Output** : target folder is created along with customer name and tenant name as mentioned in the environment variables. This folder will contain all terraform projects as mentioned below.
  
    ```
    ├── target                   # Target folder for terraform code
    │   ├── customer-name        # Folder with customer name
    │     ├── tenant-name        # Folder with tenant name
    │       ├── scripts          # Wrapper scripts to plan, apply and destroy terarform infrastructure.
    │       ├── terraform        # Terraform code generated using this utility.
    │          ├── admin-tenant  # Terraform code for tenant and tenant related resources.
    │          ├── aws-services  # Terraform code for AWS services.
    │          ├── app           # Terraform code for duplo services and ecs.
    ```

  - **Project : admin-tenant** This projects manages creation of duplo tenant and tenant related resources.
  - **Project : aws-services** This project manages data services like Redis, RDS, Kafka, S3 buckets, Cloudfront, EMR, Elastic Search inside duplo.
  - **Project : app** This project manages duplo services like eks and ecs etc.

## Following DuploCloud resources are supported.
   - `duplocloud_tenant`
   - `duplocloud_tenant_network_security_rule`
   - `duplocloud_asg_profile`
   - `duplocloud_aws_host`
   - `duplocloud_aws_kafka_cluster`
   - `duplocloud_rds_instance`
   - `duplocloud_ecache_instance`
   - `duplocloud_s3_bucket`
   - `duplocloud_aws_sns_topic`
   - `duplocloud_aws_sqs_queue`
   - `duplocloud_duplo_service`
   - `duplocloud_duplo_service_lbconfigs`
   - `duplocloud_duplo_service_params`
   - `duplocloud_ecs_task_definition`
   - `duplocloud_ecs_service`
   - `duplocloud_aws_mwaa_environment`
   - `duplocloud_aws_elasticsearch`
   - `duplocloud_k8_secret`
   - `duplocloud_k8_config_map`
   - `duplocloud_aws_ssm_parameter`
   - `duplocloud_aws_load_balancer`
   - `duplocloud_aws_load_balancer_listener`
   - `duplocloud_aws_api_gateway_integration`
   - `duplocloud_aws_ecr_repository`
   - `duplocloud_aws_cloudfront_distribution`
   - `duplocloud_aws_lambda_function`
   - `duplocloud_aws_lambda_permission`
   - `duplocloud_aws_dynamodb_table_v2`
   - `duplocloud_byoh`
   - `duplocloud_emr_cluster`
   - `duplocloud_aws_cloudwatch_metric_alarm`
   - `duplocloud_aws_cloudwatch_event_rule`
   - `duplocloud_aws_cloudwatch_event_target`
   - `duplocloud_aws_target_group_attributes`

## How to use generated terraform code to create a new DuploCloud Tenant, and its resources?

### Prerequisite
1. Following environment variables to be exported in the shell while running this terraform projects.
```shell
export AWS_RUNNER=duplo-admin
export tenant_id="XXXXXXXXXXXXXXXXXXXXXXXXX"
export duplo_host="https://msp.duplocloud.net"
export duplo_token="<duplo-auth-token>"
export aws_account_id="1234567890"
```
2. To run terraform projects you must be in `tenant-name` directory.
```shell
cd target/customer-name/tenant-name
```

### Wrapper Scripts

There are scripts to manage terraform infrastructure. Which will helps to create Duplo infrastructure based on tenant.

- scripts/plan.sh
- scripts/apply.sh
- scripts/destroy.sh

#### Arguments to run the scripts.

- **First Argument:** Name of the new tenant to be created.
- **Second Argument:** Terraform project name. Valid values are - `admin-tenant`, `aws-services` and `app`.

### Terraform Projects

This infrastructure is divided into terraform sub projects which manages different duplo resources like tenant, AWS services like Redis, RDS, Kafka, S3 buckets, Elastic Search and Duplo services which are containerized.

- **Project - admin-tenant**

  This projects manages duplo infrastructure and tenant, Run this project using following command using tenant-name and project name.

  - Dry-run

    - ```shell
       scripts/plan.sh <tenant-name> admin-tenant
      ```

  - Actual Deployment

    - ```shell
      scripts/apply.sh <tenant-name> admin-tenant
      ```

  - Destroy created infrastructure

    - ```shell
      scripts/destroy.sh <tenant-name> admin-tenant
      ```

- **Project - aws-services**

  This project manages AWS services like Redis, RDS, Kafka, S3 buckets, Elastic Search, etc. inside duplo.

  - Dry-run

    - ```shell
       scripts/plan.sh <tenant-name> aws-services
      ```
  - Actual Deployment

    - ```shell
      scripts/apply.sh <tenant-name> aws-services
      ```
  - Destroy created infrastructure

    - ```shell
      scripts/destroy.sh <tenant-name> aws-services
      ```

- **Project - app**

  This project manages containerized applications inside duplo liek eks services, ecs, docker native service etc.

  - Dry-run

    - ```shell
       scripts/plan.sh <tenant-name> app
      ```

  - Actual Deployment

    - ```shell
      scripts/apply.sh <tenant-name> app
      ```

  - Destroy created infrastructure

    - ```shell
      scripts/destroy.sh <tenant-name> app
      ```

### External Configurations (Manual Step)

End user can pass external configurations like RDS instance type, ES version, Docker image version etc. while running these projects.

#### How to create configuration file

- Create configuration folder inside config folder with the name of tenant, Example, Tenant Name is `dev01` then `dev01` folder is created inside config folder first.

- File - **admin-tenant.tfvars.json**
  - This file is used while running **Project - admin-tenant**, You can create file **admin-tenant.tfvars.json** and pass required configuration.
- File - **aws-services.tfvars.json**
  - This file is used while running **Project - aws-services**, You can create file **aws-services.tfvars.json** and pass required configuration.

- File - **app.tfvars.json**
  - This file is used while running **Project - app**, You can create file **app.tfvars.json** and pass required configuration.

    ```
    ├── target                                  # Target folder for terraform code
    │   ├── customer-name                       # Folder with customer name
    │     ├── tenant-name                       # Folder with tenant name
    │       ├── config                          # External configuration folder.
    │          ├── dev01                        # Tenant specific config folder.
    │             ├── admin-tenant.tfvars.json  # admin-tenant project variables.
    │             ├── aws-services.tfvars.json  # aws-services project variables.
    │             ├── app.tfvars.json           # app project variables.
    ```  