param location string = resourceGroup().location
param prefix string = 'cinereserve'

module postgres './modules/postgres.bicep' = {
  name: 'postgres'
  params: {
    location: location
    prefix: prefix
  }
}

module registry './modules/container-registry.bicep' = {
  name: 'registry'
  params: {
    location: location
    prefix: prefix
  }
}

module storage './modules/storage.bicep' = {
  name: 'storage'
  params: {
    location: location
    prefix: prefix
  }
}

module monitoring './modules/monitoring.bicep' = {
  name: 'monitoring'
  params: {
    location: location
    prefix: prefix
  }
}

module keyvault './modules/keyvault.bicep' = {
  name: 'keyvault'
  params: {
    location: location
    prefix: prefix
  }
}

