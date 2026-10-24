# The frontend's hosting as code: a resource group and a storage account that serves web/dist as a static website.
# Same as clicking it in the portal, but repeatable: terraform apply creates it, terraform destroy deletes it.

terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm" # plugin that talks to the Azure API
      version = "~> 4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
}

resource "azurerm_resource_group" "web" {
  name     = var.resource_group_name
  location = var.location
}

# Storage account names are global across Azure (they become <name>.blob.core.windows.net),
# so a random suffix avoids clashing with everyone else running this.
resource "random_string" "suffix" {
  length  = 6
  upper   = false
  special = false
}

resource "azurerm_storage_account" "web" {
  name                     = "stseminarweb${random_string.suffix.result}" # 3-24 lowercase letters and digits
  resource_group_name      = azurerm_resource_group.web.name              # a reference: Terraform creates the group first
  location                 = azurerm_resource_group.web.location
  account_tier             = "Standard"
  account_replication_type = "LRS" # cheapest: 3 copies in one datacenter
}

# Turns on the static website: files in the "$web" container are served over HTTPS.
resource "azurerm_storage_account_static_website" "web" {
  storage_account_id = azurerm_storage_account.web.id
  index_document     = "index.html"
  error_404_document = "index.html"
}
