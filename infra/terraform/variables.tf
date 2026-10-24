# Inputs. Set them in terraform.tfvars (see terraform.tfvars.example), with -var on the command line,
# or as environment variables named TF_VAR_<name>.

variable "subscription_id" {
  description = "Azure subscription to create the resources in (az account show --query id -o tsv)"
  type        = string
}

variable "location" {
  description = "Azure region. Student subscriptions allow only some regions: pick another on a policy error."
  type        = string
  default     = "eastus2"
}

variable "resource_group_name" {
  description = "Resource group for the frontend. terraform destroy deletes it with everything inside."
  type        = string
  default     = "rg-seminar-web"
}
