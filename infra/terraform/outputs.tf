# Printed after terraform apply; read them again any time with: terraform output

output "website_url" {
  description = "Public URL of the frontend"
  value       = azurerm_storage_account.web.primary_web_endpoint
}

output "storage_account_name" {
  description = "Needed to upload the files (az storage blob upload-batch --account-name ...)"
  value       = azurerm_storage_account.web.name
}
