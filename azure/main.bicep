param location string = resourceGroup().location
param prefix string = 'cinereserve'
@secure()
param postgresAdminPassword string
param postgresAdminLogin string = 'postgresadmin'
param storageContainerName string = 'movie-posters'
param storageQueueName string = 'ticket-emails'
param deployContainerRegistry bool = false

module postgres './modules/postgres.bicep' = {
  name: 'postgres'
  params: {
    location: location
    prefix: prefix
    administratorLogin: postgresAdminLogin
    administratorLoginPassword: postgresAdminPassword
  }
}

module registry './modules/container-registry.bicep' = if (deployContainerRegistry) {
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
    blobContainerName: storageContainerName
    queueName: storageQueueName
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
