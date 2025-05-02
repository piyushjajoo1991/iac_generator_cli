# IaC Manifest Generator CLI Examples

This document provides detailed examples of using the IaC Manifest Generator CLI to generate infrastructure as code manifests for various AWS infrastructure patterns.

## Table of Contents

- [Basic AWS Infrastructure](#basic-aws-infrastructure)
  - [Single EC2 Instance](#single-ec2-instance)
  - [S3 Bucket Configuration](#s3-bucket-configuration)
- [Networking Infrastructure](#networking-infrastructure)
  - [VPC with Public and Private Subnets](#vpc-with-public-and-private-subnets)
  - [Multi-AZ Network Configuration](#multi-az-network-configuration)
- [Compute Infrastructure](#compute-infrastructure)
  - [EKS Cluster Setup](#eks-cluster-setup)
  - [Auto-scaling EC2 Instances](#auto-scaling-ec2-instances)
- [Database Infrastructure](#database-infrastructure)
  - [RDS Database Configuration](#rds-database-configuration)
  - [DynamoDB Table Setup](#dynamodb-table-setup)
- [Complex Scenarios](#complex-scenarios)
  - [Multi-tier Web Application](#multi-tier-web-application)
  - [Microservices Infrastructure](#microservices-infrastructure)
- [Using Different Output Formats](#using-different-output-formats)
  - [Terraform vs Crossplane Comparison](#terraform-vs-crossplane-comparison)
  - [Template Customization](#template-customization)

## Basic AWS Infrastructure

### Single EC2 Instance

This example demonstrates how to generate a configuration for a simple EC2 instance.

#### Natural Language Description

```
Create an EC2 instance with t2.micro instance type in the us-east-1 region. Use the latest Amazon Linux 2 AMI and enable public IP address assignment.
```

#### Command to Generate

```bash
iacgen generate "Create an EC2 instance with t2.micro instance type in the us-east-1 region. Use the latest Amazon Linux 2 AMI and enable public IP address assignment."
```

#### Generated Terraform Output

```hcl
resource "aws_instance" "example_instance" {
  ami           = "ami-0c02fb55956c7d316" # Amazon Linux 2 AMI in us-east-1
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.public.id
  
  associate_public_ip_address = true
  
  tags = {
    Name = "example-instance"
  }
}
```

#### Notes

- The AMI ID is automatically selected based on the region and OS description
- A default name is assigned to the instance
- The instance is placed in a public subnet for public IP assignment

### S3 Bucket Configuration

This example shows how to generate a configuration for an S3 bucket with specific settings.

#### Natural Language Description

```
Create an S3 bucket named 'my-application-assets' with versioning enabled and server-side encryption. The bucket should allow public read access for website hosting.
```

#### Command to Generate

```bash
iacgen generate "Create an S3 bucket named 'my-application-assets' with versioning enabled and server-side encryption. The bucket should allow public read access for website hosting."
```

#### Generated Terraform Output

```hcl
resource "aws_s3_bucket" "my_application_assets" {
  bucket = "my-application-assets"
  
  tags = {
    Name = "my-application-assets"
  }
}

resource "aws_s3_bucket_versioning" "my_application_assets_versioning" {
  bucket = aws_s3_bucket.my_application_assets.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "my_application_assets_encryption" {
  bucket = aws_s3_bucket.my_application_assets.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_website_configuration" "my_application_assets_website" {
  bucket = aws_s3_bucket.my_application_assets.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }
}

resource "aws_s3_bucket_public_access_block" "my_application_assets_access" {
  bucket = aws_s3_bucket.my_application_assets.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "my_application_assets_policy" {
  bucket = aws_s3_bucket.my_application_assets.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.my_application_assets.arn}/*"
      }
    ]
  })
}
```

#### Notes

- The bucket name is used as specified in the description
- Multiple resources are created to configure different aspects of the S3 bucket
- Website hosting is configured with default index and error documents
- Public read access is enabled through a bucket policy

## Networking Infrastructure

### VPC with Public and Private Subnets

This example demonstrates how to generate a configuration for a VPC with both public and private subnets.

#### Natural Language Description

```
Create a VPC with CIDR block 10.0.0.0/16 in us-east-1. Add 2 public subnets and 2 private subnets. Include an internet gateway for the public subnets and a NAT gateway for the private subnets.
```

#### Command to Generate

```bash
iacgen generate "Create a VPC with CIDR block 10.0.0.0/16 in us-east-1. Add 2 public subnets and 2 private subnets. Include an internet gateway for the public subnets and a NAT gateway for the private subnets."
```

#### Generated Terraform Output

```hcl
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "main-vpc"
  }
}

resource "aws_subnet" "public" {
  count = 2

  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)
  availability_zone       = element(["us-east-1a", "us-east-1b"], count.index)
  map_public_ip_on_launch = true

  tags = {
    Name = "public-subnet-${count.index + 1}"
  }
}

resource "aws_subnet" "private" {
  count = 2

  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index + 10)
  availability_zone       = element(["us-east-1a", "us-east-1b"], count.index)
  map_public_ip_on_launch = false

  tags = {
    Name = "private-subnet-${count.index + 1}"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "main-igw"
  }
}

resource "aws_eip" "nat" {
  domain = "vpc"

  tags = {
    Name = "nat-eip"
  }
}

resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[0].id

  tags = {
    Name = "main-nat-gateway"
  }

  depends_on = [aws_internet_gateway.main]
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "public-route-table"
  }
}

resource "aws_route" "public_internet_gateway" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.main.id
}

resource "aws_route_table_association" "public" {
  count = 2

  subnet_id      = element(aws_subnet.public.*.id, count.index)
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "private-route-table"
  }
}

resource "aws_route" "private_nat_gateway" {
  route_table_id         = aws_route_table.private.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.main.id
}

resource "aws_route_table_association" "private" {
  count = 2

  subnet_id      = element(aws_subnet.private.*.id, count.index)
  route_table_id = aws_route_table.private.id
}
```

#### Notes

- The VPC CIDR block is set as specified
- Subnets are created with calculated CIDR blocks
- A single NAT gateway is created to save costs (deployed in the first public subnet)
- Route tables are configured to direct traffic appropriately

### Multi-AZ Network Configuration

This example demonstrates how to generate a highly available network architecture spanning multiple availability zones.

#### Natural Language Description

```
Create a highly available VPC architecture in us-west-2 with CIDR 172.16.0.0/16. Deploy 3 public and 3 private subnets across all availability zones. Include an internet gateway and 3 NAT gateways (one in each AZ) for fault tolerance.
```

#### Command to Generate

```bash
iacgen generate "Create a highly available VPC architecture in us-west-2 with CIDR 172.16.0.0/16. Deploy 3 public and 3 private subnets across all availability zones. Include an internet gateway and 3 NAT gateways (one in each AZ) for fault tolerance."
```

#### Generated Terraform Output

```hcl
resource "aws_vpc" "main" {
  cidr_block           = "172.16.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "multi-az-vpc"
  }
}

resource "aws_subnet" "public" {
  count = 3

  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)
  availability_zone       = element(["us-west-2a", "us-west-2b", "us-west-2c"], count.index)
  map_public_ip_on_launch = true

  tags = {
    Name = "public-subnet-${element(["us-west-2a", "us-west-2b", "us-west-2c"], count.index)}"
  }
}

resource "aws_subnet" "private" {
  count = 3

  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index + 10)
  availability_zone       = element(["us-west-2a", "us-west-2b", "us-west-2c"], count.index)
  map_public_ip_on_launch = false

  tags = {
    Name = "private-subnet-${element(["us-west-2a", "us-west-2b", "us-west-2c"], count.index)}"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "main-igw"
  }
}

resource "aws_eip" "nat" {
  count = 3
  
  domain = "vpc"

  tags = {
    Name = "nat-eip-${count.index + 1}"
  }
}

resource "aws_nat_gateway" "main" {
  count = 3

  allocation_id = element(aws_eip.nat.*.id, count.index)
  subnet_id     = element(aws_subnet.public.*.id, count.index)

  tags = {
    Name = "nat-gateway-${element(["us-west-2a", "us-west-2b", "us-west-2c"], count.index)}"
  }

  depends_on = [aws_internet_gateway.main]
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "public-route-table"
  }
}

resource "aws_route" "public_internet_gateway" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.main.id
}

resource "aws_route_table_association" "public" {
  count = 3

  subnet_id      = element(aws_subnet.public.*.id, count.index)
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "private" {
  count = 3

  vpc_id = aws_vpc.main.id

  tags = {
    Name = "private-route-table-${element(["us-west-2a", "us-west-2b", "us-west-2c"], count.index)}"
  }
}

resource "aws_route" "private_nat_gateway" {
  count = 3

  route_table_id         = element(aws_route_table.private.*.id, count.index)
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = element(aws_nat_gateway.main.*.id, count.index)
}

resource "aws_route_table_association" "private" {
  count = 3

  subnet_id      = element(aws_subnet.private.*.id, count.index)
  route_table_id = element(aws_route_table.private.*.id, count.index)
}
```

#### Notes

- The architecture spans all three availability zones in the us-west-2 region
- Each private subnet has its own NAT gateway for high availability
- Each NAT gateway is placed in a public subnet in the same availability zone
- Separate route tables are created for each private subnet

## Compute Infrastructure

### EKS Cluster Setup

This example demonstrates how to generate a configuration for an Amazon EKS cluster.

#### Natural Language Description

```
Create an EKS cluster named 'production' in us-east-2 with Kubernetes version 1.27. Deploy it in private subnets with a node group of 3 t3.large instances that can scale up to 10 nodes.
```

#### Command to Generate

```bash
iacgen generate "Create an EKS cluster named 'production' in us-east-2 with Kubernetes version 1.27. Deploy it in private subnets with a node group of 3 t3.large instances that can scale up to 10 nodes."
```

#### Generated Terraform Output

```hcl
resource "aws_eks_cluster" "production" {
  name     = "production"
  role_arn = aws_iam_role.cluster.arn
  version  = "1.27"

  vpc_config {
    subnet_ids              = aws_subnet.private.*.id
    endpoint_private_access = true
    endpoint_public_access  = true
    security_group_ids      = [aws_security_group.cluster.id]
  }

  depends_on = [
    aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy,
    aws_iam_role_policy_attachment.cluster_AmazonEKSVPCResourceController,
  ]

  tags = {
    Name = "production"
  }
}

resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.production.name
  node_group_name = "main"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.private.*.id

  instance_types = ["t3.large"]
  capacity_type  = "ON_DEMAND"
  disk_size      = 20

  scaling_config {
    desired_size = 3
    min_size     = 1
    max_size     = 10
  }

  update_config {
    max_unavailable = 1
  }

  depends_on = [
    aws_iam_role_policy_attachment.node_AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.node_AmazonEKS_CNI_Policy,
    aws_iam_role_policy_attachment.node_AmazonEC2ContainerRegistryReadOnly,
  ]

  tags = {
    Name = "production-main-ng"
  }
}

resource "aws_iam_role" "cluster" {
  name = "production-cluster-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role_policy_attachment" "cluster_AmazonEKSVPCResourceController" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role" "node" {
  name = "production-node-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "node_AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node_AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node_AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.node.name
}

resource "aws_security_group" "cluster" {
  name        = "production-cluster-sg"
  description = "Security group for EKS cluster"
  vpc_id      = aws_vpc.main.id
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "production-cluster-sg"
  }
}
```

#### Notes

- The EKS cluster is created with the specified name and version
- IAM roles and policies are created for the cluster and node group
- The node group is configured with the specified instance type and scaling parameters
- A security group is created for the cluster

### Auto-scaling EC2 Instances

This example demonstrates how to generate a configuration for auto-scaling EC2 instances.

#### Natural Language Description

```
Create an auto-scaling group of EC2 instances in us-west-1, using t3.medium instances with Amazon Linux 2. Configure scaling between 2 and 10 instances based on CPU utilization. Place instances in private subnets behind an application load balancer in public subnets.
```

#### Command to Generate

```bash
iacgen generate "Create an auto-scaling group of EC2 instances in us-west-1, using t3.medium instances with Amazon Linux 2. Configure scaling between 2 and 10 instances based on CPU utilization. Place instances in private subnets behind an application load balancer in public subnets."
```

#### Generated Terraform Output

```hcl
resource "aws_launch_template" "app" {
  name_prefix   = "app-"
  image_id      = "ami-0d382e80be7ffdae5" # Amazon Linux 2 in us-west-1
  instance_type = "t3.medium"
  
  network_interfaces {
    associate_public_ip_address = false
    security_groups             = [aws_security_group.instance.id]
  }
  
  user_data = base64encode(<<-EOF
    #!/bin/bash
    echo "Hello from user data!"
    yum update -y
    EOF
  )
  
  tag_specifications {
    resource_type = "instance"
    tags = {
      Name = "app-instance"
    }
  }
}

resource "aws_autoscaling_group" "app" {
  name                = "app-asg"
  vpc_zone_identifier = aws_subnet.private.*.id
  desired_capacity    = 2
  min_size            = 2
  max_size            = 10
  
  launch_template {
    id      = aws_launch_template.app.id
    version = "$Latest"
  }
  
  target_group_arns = [aws_lb_target_group.app.arn]
  
  health_check_type         = "ELB"
  health_check_grace_period = 300
  
  tag {
    key                 = "Name"
    value               = "app-asg-instance"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_policy" "scale_up" {
  name                   = "scale-up"
  autoscaling_group_name = aws_autoscaling_group.app.name
  adjustment_type        = "ChangeInCapacity"
  scaling_adjustment     = 1
  cooldown               = 300
}

resource "aws_cloudwatch_metric_alarm" "high_cpu" {
  alarm_name          = "high-cpu-utilization"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80
  
  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.app.name
  }
  
  alarm_description = "Scale up when CPU exceeds 80%"
  alarm_actions     = [aws_autoscaling_policy.scale_up.arn]
}

resource "aws_autoscaling_policy" "scale_down" {
  name                   = "scale-down"
  autoscaling_group_name = aws_autoscaling_group.app.name
  adjustment_type        = "ChangeInCapacity"
  scaling_adjustment     = -1
  cooldown               = 300
}

resource "aws_cloudwatch_metric_alarm" "low_cpu" {
  alarm_name          = "low-cpu-utilization"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 20
  
  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.app.name
  }
  
  alarm_description = "Scale down when CPU falls below 20%"
  alarm_actions     = [aws_autoscaling_policy.scale_down.arn]
}

resource "aws_lb" "app" {
  name               = "app-lb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.lb.id]
  subnets            = aws_subnet.public.*.id
  
  enable_deletion_protection = false
  
  tags = {
    Name = "app-lb"
  }
}

resource "aws_lb_target_group" "app" {
  name     = "app-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id
  
  health_check {
    path                = "/"
    port                = "traffic-port"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    interval            = 30
  }
}

resource "aws_lb_listener" "app" {
  load_balancer_arn = aws_lb.app.arn
  port              = 80
  protocol          = "HTTP"
  
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}

resource "aws_security_group" "lb" {
  name        = "app-lb-sg"
  description = "Security group for application load balancer"
  vpc_id      = aws_vpc.main.id
  
  ingress {
    description = "HTTP from internet"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Name = "app-lb-sg"
  }
}

resource "aws_security_group" "instance" {
  name        = "app-instance-sg"
  description = "Security group for EC2 instances"
  vpc_id      = aws_vpc.main.id
  
  ingress {
    description     = "HTTP from load balancer"
    from_port       = 80
    to_port         = 80
    protocol        = "tcp"
    security_groups = [aws_security_group.lb.id]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Name = "app-instance-sg"
  }
}
```

#### Notes

- A launch template is used instead of a launch configuration for newer features
- Auto-scaling policies and CloudWatch alarms are created for scaling based on CPU
- An Application Load Balancer is placed in public subnets with a target group
- Security groups are configured to allow traffic from the load balancer to instances

## Database Infrastructure

### RDS Database Configuration

This example demonstrates how to generate a configuration for an Amazon RDS database.

#### Natural Language Description

```
Create a highly available MySQL RDS database in us-east-1 with multi-AZ deployment. Use db.t3.medium instance class with 100GB storage that can automatically scale up to 1TB. Place the database in private subnets and secure it with a database security group.
```

#### Command to Generate

```bash
iacgen generate "Create a highly available MySQL RDS database in us-east-1 with multi-AZ deployment. Use db.t3.medium instance class with 100GB storage that can automatically scale up to 1TB. Place the database in private subnets and secure it with a database security group."
```

#### Generated Terraform Output

```hcl
resource "aws_db_subnet_group" "default" {
  name       = "main-db-subnet-group"
  subnet_ids = aws_subnet.private.*.id

  tags = {
    Name = "Main DB subnet group"
  }
}

resource "aws_security_group" "db" {
  name        = "database-sg"
  description = "Security group for database instances"
  vpc_id      = aws_vpc.main.id

  ingress {
    description     = "MySQL access from app instances"
    from_port       = 3306
    to_port         = 3306
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "database-sg"
  }
}

resource "aws_db_parameter_group" "mysql" {
  name   = "mysql-parameters"
  family = "mysql8.0"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }
}

resource "aws_db_instance" "mysql" {
  identifier           = "main-mysql-db"
  engine               = "mysql"
  engine_version       = "8.0"
  instance_class       = "db.t3.medium"
  allocated_storage    = 100
  max_allocated_storage = 1024
  storage_type         = "gp2"
  
  username             = "admin"
  password             = "password"  # In production, use AWS Secrets Manager
  
  db_subnet_group_name = aws_db_subnet_group.default.name
  vpc_security_group_ids = [aws_security_group.db.id]
  
  multi_az             = true
  publicly_accessible  = false
  skip_final_snapshot  = true
  
  parameter_group_name = aws_db_parameter_group.mysql.name
  
  backup_retention_period = 7
  backup_window           = "03:00-04:00"
  maintenance_window      = "Mon:00:00-Mon:03:00"
  
  tags = {
    Name = "main-mysql-db"
  }
}

resource "aws_security_group" "app" {
  name        = "app-sg"
  description = "Security group for application instances"
  vpc_id      = aws_vpc.main.id
  
  # This is just a placeholder for the reference in the db security group
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Name = "app-sg"
  }
}
```

#### Notes

- The RDS instance is configured for multi-AZ deployment for high availability
- The database is placed in private subnets using a subnet group
- Storage autoscaling is enabled with the specified limits
- A security group is created to control access to the database
- A parameter group is created with UTF-8 character set settings

### DynamoDB Table Setup

This example demonstrates how to generate a configuration for an Amazon DynamoDB table.

#### Natural Language Description

```
Create a DynamoDB table named 'users' with a partition key of 'user_id' (string) and a sort key of 'email' (string). Configure on-demand capacity mode, enable point-in-time recovery, and add a global secondary index on the 'status' attribute.
```

#### Command to Generate

```bash
iacgen generate "Create a DynamoDB table named 'users' with a partition key of 'user_id' (string) and a sort key of 'email' (string). Configure on-demand capacity mode, enable point-in-time recovery, and add a global secondary index on the 'status' attribute."
```

#### Generated Terraform Output

```hcl
resource "aws_dynamodb_table" "users" {
  name           = "users"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "user_id"
  range_key      = "email"

  attribute {
    name = "user_id"
    type = "S"
  }

  attribute {
    name = "email"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  global_secondary_index {
    name            = "status-index"
    hash_key        = "status"
    projection_type = "ALL"
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Name = "users-table"
  }
}
```

#### Notes

- The table is configured with the specified partition and sort keys
- On-demand (pay per request) billing mode is enabled
- Point-in-time recovery is enabled for backup protection
- A global secondary index is created on the status attribute
- All required attributes are defined

## Complex Scenarios

### Multi-tier Web Application

This example demonstrates how to generate a configuration for a multi-tier web application.

#### Natural Language Description

```
Create a 3-tier web application infrastructure in us-east-1 with:
- A VPC with 3 public and 3 private subnets across all availability zones
- An internet-facing application load balancer in the public subnets
- An auto-scaling group of web servers using t3.medium instances in the private subnets
- A MySQL RDS database with multi-AZ deployment
- An ElastiCache Redis cluster for session caching
- An S3 bucket for static assets with CloudFront distribution
```

#### Command to Generate

```bash
iacgen generate "Create a 3-tier web application infrastructure in us-east-1 with:
- A VPC with 3 public and 3 private subnets across all availability zones
- An internet-facing application load balancer in the public subnets
- An auto-scaling group of web servers using t3.medium instances in the private subnets
- A MySQL RDS database with multi-AZ deployment
- An ElastiCache Redis cluster for session caching
- An S3 bucket for static assets with CloudFront distribution"
```

#### Generated Terraform Output (Partial)

Below is a partial output for the multi-tier web application. For brevity, only key resources are shown:

```hcl
# VPC Configuration
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "web-app-vpc"
  }
}

# Public Subnets
resource "aws_subnet" "public" {
  count = 3

  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)
  availability_zone       = element(["us-east-1a", "us-east-1b", "us-east-1c"], count.index)
  map_public_ip_on_launch = true

  tags = {
    Name = "public-subnet-${count.index + 1}"
  }
}

# Private Subnets
resource "aws_subnet" "private" {
  count = 3

  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(aws_vpc.main.cidr_block, 8, count.index + 10)
  availability_zone       = element(["us-east-1a", "us-east-1b", "us-east-1c"], count.index)
  map_public_ip_on_launch = false

  tags = {
    Name = "private-subnet-${count.index + 1}"
  }
}

# Application Load Balancer
resource "aws_lb" "web" {
  name               = "web-app-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id

  enable_deletion_protection = false

  tags = {
    Name = "web-app-alb"
  }
}

# Auto-scaling Group
resource "aws_launch_template" "web" {
  name_prefix   = "web-app-"
  image_id      = "ami-0c55b159cbfafe1f0" # Amazon Linux 2 in us-east-1
  instance_type = "t3.medium"

  network_interfaces {
    associate_public_ip_address = false
    security_groups             = [aws_security_group.web.id]
  }
  
  user_data = base64encode(<<-EOF
    #!/bin/bash
    yum update -y
    yum install -y httpd php
    systemctl start httpd
    systemctl enable httpd
  EOF
  )
}

resource "aws_autoscaling_group" "web" {
  name                = "web-app-asg"
  vpc_zone_identifier = aws_subnet.private[*].id
  desired_capacity    = 3
  min_size            = 2
  max_size            = 10
  
  launch_template {
    id      = aws_launch_template.web.id
    version = "$Latest"
  }
  
  target_group_arns = [aws_lb_target_group.web.arn]
}

# RDS Database
resource "aws_db_subnet_group" "default" {
  name       = "web-app-db-subnet-group"
  subnet_ids = aws_subnet.private[*].id
}

resource "aws_db_instance" "mysql" {
  identifier           = "web-app-db"
  engine               = "mysql"
  engine_version       = "8.0"
  instance_class       = "db.t3.medium"
  allocated_storage    = 100
  multi_az             = true
  db_subnet_group_name = aws_db_subnet_group.default.name
  vpc_security_group_ids = [aws_security_group.db.id]
  
  username             = "admin"
  password             = "password" # Use AWS Secrets Manager in production
  
  skip_final_snapshot  = true
}

# ElastiCache Redis
resource "aws_elasticache_subnet_group" "default" {
  name       = "web-app-cache-subnet-group"
  subnet_ids = aws_subnet.private[*].id
}

resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "web-app-redis"
  engine               = "redis"
  node_type            = "cache.t3.medium"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis6.x"
  subnet_group_name    = aws_elasticache_subnet_group.default.name
  security_group_ids   = [aws_security_group.redis.id]
}

# S3 Bucket for Static Assets
resource "aws_s3_bucket" "assets" {
  bucket = "web-app-static-assets"
  
  tags = {
    Name = "web-app-static-assets"
  }
}

# CloudFront Distribution
resource "aws_cloudfront_distribution" "assets" {
  origin {
    domain_name = aws_s3_bucket.assets.bucket_regional_domain_name
    origin_id   = "S3-web-app-static-assets"
  }
  
  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"
  
  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "S3-web-app-static-assets"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }
  
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }
  
  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
```

The complete configuration would also include:

- Security groups for each component
- IAM roles for services
- Route tables and network configurations
- CloudWatch alarms and scaling policies
- Detailed service configurations

### Microservices Infrastructure

This example demonstrates how to generate a configuration for a microservices architecture on AWS.

#### Natural Language Description

```
Create a microservices infrastructure in us-west-2 with:
- An EKS cluster version 1.27 with 3 node groups (general, compute-optimized, and memory-optimized)
- A VPC with private subnets for the EKS cluster
- An AWS Load Balancer Controller for ingress
- An ECR repository for container images
- A CloudWatch log group for centralized logging
- An ElastiCache Redis cluster for caching
- An RDS Aurora PostgreSQL cluster for database
```

#### Command to Generate

```bash
iacgen generate "Create a microservices infrastructure in us-west-2 with:
- An EKS cluster version 1.27 with 3 node groups (general, compute-optimized, and memory-optimized)
- A VPC with private subnets for the EKS cluster
- An AWS Load Balancer Controller for ingress
- An ECR repository for container images
- A CloudWatch log group for centralized logging
- An ElastiCache Redis cluster for caching
- An RDS Aurora PostgreSQL cluster for database"
```

#### Generated Crossplane Output (Partial)

Below is a partial output showing the Crossplane YAML for some of the key components:

```yaml
# EKS Cluster Definition
apiVersion: eks.aws.crossplane.io/v1beta1
kind: Cluster
metadata:
  name: microservices-cluster
spec:
  forProvider:
    region: us-west-2
    version: "1.27"
    roleArnSelector:
      matchControllerRef: true
    resourcesVpcConfig:
      endpointPrivateAccess: true
      endpointPublicAccess: true
      subnetIdSelector:
        matchLabels:
          type: private
  providerConfigRef:
    name: aws-provider
---
# General Node Group
apiVersion: eks.aws.crossplane.io/v1alpha1
kind: NodeGroup
metadata:
  name: general-nodes
spec:
  forProvider:
    region: us-west-2
    clusterNameSelector:
      matchControllerRef: true
    nodeRoleArnSelector:
      matchControllerRef: true
    subnetIdSelector:
      matchLabels:
        type: private
    scalingConfig:
      desiredSize: 3
      minSize: 1
      maxSize: 5
    instanceTypes:
      - t3.large
    capacityType: ON_DEMAND
  providerConfigRef:
    name: aws-provider
---
# Compute-Optimized Node Group
apiVersion: eks.aws.crossplane.io/v1alpha1
kind: NodeGroup
metadata:
  name: compute-nodes
spec:
  forProvider:
    region: us-west-2
    clusterNameSelector:
      matchControllerRef: true
    nodeRoleArnSelector:
      matchControllerRef: true
    subnetIdSelector:
      matchLabels:
        type: private
    scalingConfig:
      desiredSize: 2
      minSize: 1
      maxSize: 10
    instanceTypes:
      - c5.xlarge
    capacityType: ON_DEMAND
  providerConfigRef:
    name: aws-provider
---
# Memory-Optimized Node Group
apiVersion: eks.aws.crossplane.io/v1alpha1
kind: NodeGroup
metadata:
  name: memory-nodes
spec:
  forProvider:
    region: us-west-2
    clusterNameSelector:
      matchControllerRef: true
    nodeRoleArnSelector:
      matchControllerRef: true
    subnetIdSelector:
      matchLabels:
        type: private
    scalingConfig:
      desiredSize: 2
      minSize: 1
      maxSize: 8
    instanceTypes:
      - r5.large
    capacityType: ON_DEMAND
  providerConfigRef:
    name: aws-provider
---
# VPC for EKS
apiVersion: ec2.aws.crossplane.io/v1beta1
kind: VPC
metadata:
  name: microservices-vpc
spec:
  forProvider:
    region: us-west-2
    cidrBlock: 10.0.0.0/16
    enableDnsHostnames: true
    enableDnsSupport: true
    tags:
      - key: Name
        value: microservices-vpc
  providerConfigRef:
    name: aws-provider
---
# ECR Repository
apiVersion: ecr.aws.crossplane.io/v1beta1
kind: Repository
metadata:
  name: microservices-repo
spec:
  forProvider:
    region: us-west-2
    imageScanningConfiguration:
      scanOnPush: true
    imageTagMutability: MUTABLE
  providerConfigRef:
    name: aws-provider
---
# CloudWatch Log Group
apiVersion: cloudwatch.aws.crossplane.io/v1alpha1
kind: LogGroup
metadata:
  name: microservices-logs
spec:
  forProvider:
    region: us-west-2
    retentionInDays: 30
  providerConfigRef:
    name: aws-provider
---
# ElastiCache Redis
apiVersion: cache.aws.crossplane.io/v1beta1
kind: ReplicationGroup
metadata:
  name: microservices-redis
spec:
  forProvider:
    region: us-west-2
    engine: redis
    cacheNodeType: cache.t3.medium
    numNodeGroups: 1
    replicasPerNodeGroup: 1
    automaticFailoverEnabled: true
    cacheSubnetGroupNameSelector:
      matchControllerRef: true
    securityGroupIdSelector:
      matchControllerRef: true
  providerConfigRef:
    name: aws-provider
---
# Aurora PostgreSQL Cluster
apiVersion: rds.aws.crossplane.io/v1beta1
kind: DBCluster
metadata:
  name: microservices-db
spec:
  forProvider:
    region: us-west-2
    engine: aurora-postgresql
    engineVersion: "13.7"
    dbSubnetGroupNameSelector:
      matchControllerRef: true
    vpcSecurityGroupIdSelector:
      matchControllerRef: true
    masterUsername: admin
    masterPasswordSecretRef:
      key: password
      name: db-password
      namespace: default
    skipFinalSnapshot: true
  providerConfigRef:
    name: aws-provider
```

The complete configuration would also include:

- IAM role definitions for EKS and services
- Subnet configurations for the VPC
- Additional security groups
- Load Balancer Controller resources
- Service accounts for Kubernetes integration

## Using Different Output Formats

### Terraform vs Crossplane Comparison

This example demonstrates the difference between Terraform and Crossplane outputs for the same infrastructure description.

#### Natural Language Description

```
Create an S3 bucket named 'example-data' with versioning enabled and server-side encryption in us-east-1 region.
```

#### Terraform Output

```hcl
resource "aws_s3_bucket" "example_data" {
  bucket = "example-data"
  
  tags = {
    Name = "example-data"
  }
}

resource "aws_s3_bucket_versioning" "example_data_versioning" {
  bucket = aws_s3_bucket.example_data.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "example_data_encryption" {
  bucket = aws_s3_bucket.example_data.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
```

#### Crossplane Output

```yaml
apiVersion: s3.aws.crossplane.io/v1beta1
kind: Bucket
metadata:
  name: example-data
spec:
  forProvider:
    region: us-east-1
    acl: private
    versioning:
      status: Enabled
    serverSideEncryptionConfiguration:
      rules:
        - applyServerSideEncryptionByDefault:
            sseAlgorithm: AES256
    tags:
      - key: Name
        value: example-data
  providerConfigRef:
    name: aws-provider
```

#### Comparison Notes

- **Format**: Terraform uses HCL syntax, while Crossplane uses YAML.
- **Resource Organization**: Terraform defines multiple separate resources, whereas Crossplane defines a single resource with nested properties.
- **References**: Terraform uses explicit references between resources, while Crossplane uses a more declarative approach with all properties defined in one place.
- **Resource Naming**: Terraform requires each resource to have a logical name, while Crossplane uses metadata.name for the resource identifier.
- **Provider Configuration**: Terraform configures providers at the module level, while Crossplane explicitly references a provider configuration for each resource.

### Template Customization

This section demonstrates using the template system to customize the output.

#### Command

```bash
iacgen generate --use-templates "Create an EC2 instance with t2.micro instance type in us-east-1."
```

Using the template system allows for customized output formats and structures based on the templates defined in the `internal/template/templates/` directory.