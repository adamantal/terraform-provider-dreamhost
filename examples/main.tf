terraform {
  required_providers {
    dreamhost = {
      version = "0.2"
      source  = "hashicorp.com/edu/dreamhost"
    }
  }
}

provider "dreamhost" {
  api_key = ""
}
