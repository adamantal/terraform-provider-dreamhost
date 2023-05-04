---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dreamhost Provider"
subcategory: ""
description: |-
  
---

# dreamhost Provider

The Dreamhost provider is used to mange Dreamhost DNS records. Users can create, delete and import DNS records using this provider.

The DNS record resources are considered immutable, so any changes in their settings replaces them (delete and create).

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `api_key` (String, Sensitive) the key to access the Dreamhost API