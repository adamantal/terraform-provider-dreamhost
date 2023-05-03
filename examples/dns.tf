resource "dreamhost_dns_record" "example_com" {
  record = "example.com"
  type   = "A"
  value  = "1.2.3.4"
}