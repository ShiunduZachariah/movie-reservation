param location string
param prefix string

resource storage 'Microsoft.Storage/storageAccounts@2023-05-01' = {
  name: '${prefix}storage'
  location: location
  sku: {
    name: 'Standard_LRS'
  }
  kind: 'StorageV2'
}

