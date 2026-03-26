param location string
param prefix string

resource server 'Microsoft.DBforPostgreSQL/flexibleServers@2023-06-01-preview' = {
  name: '${prefix}-pg'
  location: location
  sku: {
    name: 'Standard_B1ms'
    tier: 'Burstable'
  }
  properties: {
    version: '16'
    administratorLogin: 'postgresadmin'
    administratorLoginPassword: 'replace-me'
    storage: {
      storageSizeGB: 32
    }
  }
}

