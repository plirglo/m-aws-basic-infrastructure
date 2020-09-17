output "private_ip" {
  value = aws_instance.awsbi.*.private_ip
}

output "public_ip" {
  value = aws_instance.awsbi.*.public_ip
}
