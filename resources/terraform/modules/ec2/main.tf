resource "aws_resourcegroups_group" "rg" {
  name = "rg-${var.name}"

  resource_query {
    query = <<JSON
{
  "ResourceTypeFilters": [
    "AWS::DynamoDB::Table",
    "AWS::EC2::CustomerGateway",
    "AWS::EC2::DHCPOptions",
    "AWS::EC2::EIP",
    "AWS::EC2::Image",
    "AWS::EC2::Instance",
    "AWS::EC2::InternetGateway",
    "AWS::EC2::NetworkAcl",
    "AWS::EC2::NetworkInterface",
    "AWS::EC2::ReservedInstance",
    "AWS::EC2::RouteTable",
    "AWS::EC2::SecurityGroup",
    "AWS::EC2::Snapshot",
    "AWS::EC2::SpotInstanceRequest",
    "AWS::EC2::Subnet",
    "AWS::EC2::VPC",
    "AWS::EC2::VPNConnection",
    "AWS::EC2::VPNGateway",
    "AWS::EC2::Volume",
    "AWS::EMR::Cluster",
    "AWS::ElastiCache::CacheCluster",
    "AWS::ElastiCache::Snapshot",
    "AWS::ElasticLoadBalancing::LoadBalancer",
    "AWS::Glacier::Vault",
    "AWS::Kinesis::Stream",
    "AWS::Lambda::Function",
    "AWS::RDS::DBInstance",
    "AWS::RDS::DBParameterGroup",
    "AWS::RDS::DBSecurityGroup",
    "AWS::RDS::DBSnapshot",
    "AWS::RDS::DBSubnetGroup",
    "AWS::RDS::EventSubscription",
    "AWS::RDS::OptionGroup",
    "AWS::RDS::ReservedDBInstance",
    "AWS::Redshift::Cluster",
    "AWS::Redshift::ClusterParameterGroup",
    "AWS::Redshift::ClusterSubnetGroup",
    "AWS::Redshift::HSMClientCertificate",
    "AWS::ResourceGroups::Group",
    "AWS::S3::Bucket",
    "AWS::StorageGateway::Gateway"
  ],
  "TagFilters": [
    {
      "Key": "cluster_name",
      "Values": ["${var.name}"]
    }
  ]
}
JSON
  }
}

resource "aws_subnet" "awsbi-subnet" {
	vpc_id            = aws_vpc.awsbi-vpc.id
	cidr_block        = "10.1.1.0/24"
	availability_zone = "${var.region}a"
  tags              = {
    Name = "subnet-${var.name}"
  	cluster_name = var.name
  }
}

resource "aws_vpc" "awsbi-vpc" {
  cidr_block            = "10.1.0.0/20"
  instance_tenancy      = "default"
  enable_dns_support    = "true"
  enable_dns_hostnames  = "true"
  tags                  = {
    Name          = "vpc-${var.name}"
    cluster_name  = var.name
  }
}

resource "aws_internet_gateway" "awsbi-internet-gateway" {
  vpc_id  = aws_vpc.awsbi-vpc.id
  tags    = {
    Name          = "internet-gateway-${var.name}"
    cluster_name  = var.name
  }
}

resource "aws_security_group" "awsbi-security-group" {
  vpc_id  = aws_vpc.awsbi-vpc.id
  tags    = {
    Name = "sg-${var.name}"
    cluster_name = var.name
  }

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "awsbi" {
  count                       = var.instance_count
  ami                         = var.ami
  instance_type               = var.instance_type
  subnet_id                   = aws_subnet.awsbi-subnet.id
  associate_public_ip_address = var.use_public_ip
  key_name                    = var.key_name
  security_groups             = [
    aws_security_group.awsbi-security-group.id
  ]
  tags = {
    Name = var.name
  }
}

resource "aws_route_table" "awsbi-route-table" {
  vpc_id  = aws_vpc.awsbi-vpc.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.awsbi-internet-gateway.id
  }
  tags = {
        Name = "route-${var.name}"
        cluster_name = var.name
  }
}

resource "aws_route_table_association" "awsbi-route-association" {
  subnet_id      = aws_subnet.awsbi-subnet.id
  route_table_id = aws_route_table.awsbi-route-table.id
}
