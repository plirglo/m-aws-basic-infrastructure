data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "awsbi_vpc" {
  cidr_block            = var.vpc_cidr_block
  instance_tenancy      = "default"
  enable_dns_support    = "true"
  enable_dns_hostnames  = "true"

  tags = {
    name          = "${var.name}-vpc"
    cluster_name  = var.name
  }
}

resource "aws_security_group" "awsbi_security_group" {
  name    = "${var.name}-sg"
  vpc_id  = aws_vpc.awsbi_vpc.id

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

  tags = {
    cluster_name = var.name
  }
}

# --- Public ---

resource "aws_subnet" "awsbi_public_subnet" {
  count             = length(local.public_cidr_blocks)
  vpc_id            = aws_vpc.awsbi_vpc.id
  cidr_block        = local.public_cidr_blocks[count.index]
  availability_zone = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    name         = "${var.name}-subnet-public${count.index}"
    cluster_name = var.name
  }
}

resource "aws_internet_gateway" "awsbi_internet_gateway" {
  vpc_id = aws_vpc.awsbi_vpc.id

  tags = {
    name         = "${var.name}-ig"
    cluster_name = var.name
  }
}

resource "aws_route_table" "awsbi_route_table_public" {
  vpc_id  = aws_vpc.awsbi_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.awsbi_internet_gateway.id
  }

  tags = {
    name = "${var.name}-rt-public"
    cluster_name = var.name
  }
}

resource "aws_route_table_association" "awsbi_route_association_public" {
  count          = length(local.public_cidr_blocks)
  subnet_id      = aws_subnet.awsbi_public_subnet[count.index].id
  route_table_id = aws_route_table.awsbi_route_table_public.id
}

# --- Private ---

resource "aws_subnet" "awsbi_private_subnet" {
  count             = length(local.private_cidr_blocks)
  vpc_id            = aws_vpc.awsbi_vpc.id
  cidr_block        = local.private_cidr_blocks[count.index]
  availability_zone = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    name         = "${var.name}-subnet-private${count.index}"
    cluster_name = var.name
  }
}

resource "aws_eip" "awsbi_nat_gateway" {
  count = var.nat_gateway_count

  vpc = true

  tags = {
    name         = "${var.name}-eip${count.index}"
    cluster_name = var.name
  }
}

resource "aws_nat_gateway" "awsbi_nat_gateway" {
  count = var.nat_gateway_count

  allocation_id = aws_eip.awsbi_nat_gateway[count.index].id
  subnet_id     = aws_subnet.awsbi_public_subnet[count.index].id

  tags = {
    name         = "${var.name}-ng${count.index}"
    cluster_name = var.name
  }

  depends_on = [ aws_internet_gateway.awsbi_internet_gateway ]
}

resource "aws_route_table" "awsbi_route_table_private" {
  count = local.use_nat_gateway ? 1 : 0

  vpc_id = aws_vpc.awsbi_vpc.id

  dynamic "route" {
    for_each = aws_nat_gateway.awsbi_nat_gateway.*.id
    iterator = nat_gateway_id
    content {
      cidr_block     = "0.0.0.0/0"
      nat_gateway_id = nat_gateway_id.value
    }
  }

  tags = {
    name = "${var.name}-rt-private"
    cluster_name = var.name
  }
}

resource "aws_route_table_association" "awsbi_route_association_private" {
  count          = length(local.private_cidr_blocks)
  subnet_id      = aws_subnet.awsbi_private_subnet[count.index].id
  route_table_id = aws_route_table.awsbi_route_table_private[0].id
}
