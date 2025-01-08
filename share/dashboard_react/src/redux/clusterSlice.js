import { createSlice, createAsyncThunk, isAnyOf } from '@reduxjs/toolkit'
import { clusterService } from '../services/clusterService'
import { handleError, showErrorBanner, showSuccessBanner } from '../utility/common'

export const getClusterData = createAsyncThunk('cluster/getClusterData', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.getClusterData(clusterName, baseURL)
    return { data, status }
  } catch (error) {
    handleError(error, thunkAPI)
  }
})

export const getClusterAlerts = createAsyncThunk('cluster/getClusterAlerts', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.getClusterAlerts(clusterName, baseURL)
    return { data, status }
  } catch (error) {
    handleError(error, thunkAPI)
  }
})

export const getClusterMaster = createAsyncThunk('cluster/getClusterMaster', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.getClusterMaster(clusterName, baseURL)
    return { data, status }
  } catch (error) {
    handleError(error, thunkAPI)
  }
})

export const getClusterServers = createAsyncThunk('cluster/getClusterServers', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.getClusterServers(clusterName, baseURL)
    return { data, status }
  } catch (error) {
    handleError(error, thunkAPI)
  }
})

export const getClusterProxies = createAsyncThunk('cluster/getClusterProxies', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.getClusterProxies(clusterName, baseURL)
    return { data, status }
  } catch (error) {
    handleError(error, thunkAPI)
  }
})

export const getClusterCertificates = createAsyncThunk(
  'cluster/getClusterCertificates',
  async ({ clusterName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.getClusterCertificates(clusterName, baseURL)
      return { data, status }
    } catch (error) {
      handleError(error, thunkAPI)
    }
  }
)

export const getTopProcess = createAsyncThunk('cluster/getTopProcess', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.getTopProcess(clusterName, baseURL)
    return { data, status }
  } catch (error) {
    handleError(error, thunkAPI)
  }
})

export const getBackupSnapshot = createAsyncThunk('cluster/getBackupSnapshot', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.getBackupSnapshot(clusterName, baseURL)
    return { data, status }
  } catch (error) {
    handleError(error, thunkAPI)
  }
})

export const getJobs = createAsyncThunk('cluster/getJobs', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.getJobs(clusterName, baseURL)
    return { data, status }
  } catch (error) {
    handleError(error, thunkAPI)
  }
})

export const getShardSchema = createAsyncThunk('cluster/getShardSchema', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.getShardSchema(clusterName, baseURL)
    return { data, status }
  } catch (error) {
    handleError(error, thunkAPI)
  }
})

export const getQueryRules = createAsyncThunk('cluster/getQueryRules', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.getQueryRules(clusterName, baseURL)
    return { data, status }
  } catch (error) {
    handleError(error, thunkAPI)
  }
})

export const switchOverCluster = createAsyncThunk('cluster/switchOverCluster', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.switchOverCluster(clusterName, baseURL)
    showSuccessBanner('Switchover Successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Switchover Failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const failOverCluster = createAsyncThunk('cluster/failOverCluster', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.failOverCluster(clusterName, baseURL)
    showSuccessBanner('Failover Successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Failover Failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const resetFailOverCounter = createAsyncThunk(
  'cluster/resetFailOverCounter',
  async ({ clusterName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.resetFailOverCounter(clusterName, baseURL)
      showSuccessBanner('Failover counter reset!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Failover counter reset failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)
export const resetSLA = createAsyncThunk('cluster/resetSLA', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.resetSLA(clusterName, baseURL)
    showSuccessBanner('SLA reset!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('SLA reset failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const toggleTraffic = createAsyncThunk('cluster/toggleTraffic', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.toggleTraffic(clusterName, baseURL)
    showSuccessBanner('Traffic toggle done!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Traffic toggle failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const addServer = createAsyncThunk(
  'cluster/addServer',
  async ({ clusterName, host, port, dbType }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.addServer(clusterName, host, port, dbType, baseURL)
      showSuccessBanner('New server added!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Error while adding a new server', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const provisionCluster = createAsyncThunk('cluster/provisionCluster', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.provisionCluster(clusterName, baseURL)
    showSuccessBanner('Cluster provision successful', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Cluster provision failed', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const unProvisionCluster = createAsyncThunk('cluster/unProvisionCluster', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.unProvisionCluster(clusterName, baseURL)
    showSuccessBanner('Cluster unprovision successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Cluster unprovision failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const setCredentials = createAsyncThunk(
  'cluster/setCredentials',
  async ({ clusterName, credentialType, credential }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.setCredentials(clusterName, credentialType, credential, baseURL)
      showSuccessBanner(`Credentials for ${credentialType} set!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Setting credentials for ${credentialType} failed!`, error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const sendCredentials = createAsyncThunk(
  'cluster/sendCredentials',
  async ({ clusterName, username, type }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.sendCredentials(clusterName, username, type, baseURL)
      showSuccessBanner('Credentials sent to email!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Sending credentials email failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const rotateDBCredential = createAsyncThunk('cluster/rotateDBCredential', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.rotateDBCredential(clusterName, baseURL)
    showSuccessBanner('Database rotation successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Database rotation failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const rollingOptimize = createAsyncThunk('cluster/rollingOptimize', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.rollingOptimize(clusterName, baseURL)
    showSuccessBanner('Rolling optimize successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Rolling optimize failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const rollingRestart = createAsyncThunk('cluster/rollingRestart', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.rollingRestart(clusterName, baseURL)
    showSuccessBanner('Rolling restart successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Rolling restart failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const rotateCertificates = createAsyncThunk('cluster/rotateCertificates', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.rotateCertificates(clusterName, baseURL)
    showSuccessBanner('Rotate certificates successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Rotate certificates failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const reloadCertificates = createAsyncThunk('cluster/reloadCertificates', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.reloadCertificates(clusterName, baseURL)
    showSuccessBanner('Reload certificates successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Reload certificates failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const cancelRollingRestart = createAsyncThunk(
  'cluster/cancelRollingRestart',
  async ({ clusterName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.cancelRollingRestart(clusterName, baseURL)
      showSuccessBanner('Rolling restart cancelled!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Rolling restart cancellation failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const cancelRollingReprov = createAsyncThunk(
  'cluster/cancelRollingReprov',
  async ({ clusterName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.cancelRollingReprov(clusterName, baseURL)
      showSuccessBanner('Rolling reprov cancelled!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Rolling reprov cancellation failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const bootstrapMasterSlave = createAsyncThunk(
  'cluster/bootstrapMasterSlave',
  async ({ clusterName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.bootstrapMasterSlave(clusterName, baseURL)
      showSuccessBanner('Master slave bootstrap successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Master slave bootstrap failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const bootstrapMasterSlaveNoGtid = createAsyncThunk(
  'cluster/bootstrapMasterSlaveNoGtid',
  async ({ clusterName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.bootstrapMasterSlaveNoGtid(clusterName, baseURL)
      showSuccessBanner('Master slave positional bootstrap successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Master slave positional bootstrap failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const bootstrapMultiMaster = createAsyncThunk(
  'cluster/bootstrapMultiMaster',
  async ({ clusterName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.bootstrapMultiMaster(clusterName, baseURL)
      showSuccessBanner('Multi master bootstrap successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Multi master bootstrap failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const bootstrapMultiMasterRing = createAsyncThunk(
  'cluster/bootstrapMultiMasterRing',
  async ({ clusterName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.bootstrapMultiMasterRing(clusterName, baseURL)
      showSuccessBanner('Multi master ring bootstrap successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Multi master ring bootstrap failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const bootstrapMultiTierSlave = createAsyncThunk(
  'cluster/bootstrapMultiTierSlave',
  async ({ clusterName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.bootstrapMultiTierSlave(clusterName, baseURL)
      showSuccessBanner('Multitier slave bootstrap successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Multitier slave bootstrap failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const configReload = createAsyncThunk('cluster/configReload', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.configReload(clusterName, baseURL)
    showSuccessBanner('Config is reloaded!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Config reload failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const configDiscoverDB = createAsyncThunk('cluster/configDiscoverDB', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.configDiscoverDB(clusterName, baseURL)
    showSuccessBanner('Databse discover config successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Databse discover config failed!', error.message, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const configDynamic = createAsyncThunk('cluster/configDynamic', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.configDynamic(clusterName, baseURL)
    showSuccessBanner('Databse apply dynamic config successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Databse apply dynamic config failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const checksumAllTables = createAsyncThunk('cluster/checksumAllTables', async ({ clusterName }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.checksumAllTables(clusterName, baseURL)
    showSuccessBanner('Checksum all tables successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Checksum all tables failed!', error.message, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const setMaintenanceMode = createAsyncThunk(
  'cluster/setMaintenanceMode',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.setMaintenanceMode(clusterName, serverId, baseURL)
      showSuccessBanner('Maintenance mode is set!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Setting Maintenance mode failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)
export const promoteToLeader = createAsyncThunk(
  'cluster/promoteToLeader',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.promoteToLeader(clusterName, serverId, baseURL)
      showSuccessBanner('Promote to leader successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Promote to leader failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const setAsUnrated = createAsyncThunk('cluster/setAsUnrated', async ({ clusterName, serverId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.setAsUnrated(clusterName, serverId, baseURL)
    showSuccessBanner('Failover candidate set as unrated!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Failover candidate failed to set as unrated', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const setAsPreferred = createAsyncThunk(
  'cluster/setAsPreferred',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.setAsPreferred(clusterName, serverId, baseURL)
      showSuccessBanner('Failover candidate set as preferred!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Failover candidate failed to set as preferred', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const setAsIgnored = createAsyncThunk('cluster/setAsIgnored', async ({ clusterName, serverId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.setAsIgnored(clusterName, serverId, baseURL)
    showSuccessBanner('Failover candidate set as ignored!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Failover candidate failed to set as ignored', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const reseedLogicalFromBackup = createAsyncThunk(
  'cluster/reseedLogicalFromBackup',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.reseedLogicalFromBackup(clusterName, serverId, baseURL)
      showSuccessBanner('Reseed logical from backup successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Reseed logical from backup failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const reseedLogicalFromMaster = createAsyncThunk(
  'cluster/reseedLogicalFromMaster',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.reseedLogicalFromMaster(clusterName, serverId, baseURL)
      showSuccessBanner('Reseed logical from master successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Reseed logical from master failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const reseedPhysicalFromBackup = createAsyncThunk(
  'cluster/reseedPhysicalFromBackup',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.reseedPhysicalFromBackup(clusterName, serverId, baseURL)
      showSuccessBanner('Reseed physical from backup successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Reseed physical from backup failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const flushLogs = createAsyncThunk('cluster/flushLogs', async ({ clusterName, serverId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.flushLogs(clusterName, serverId, baseURL)
    showSuccessBanner('Logs flush successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Logs flush failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const physicalBackupMaster = createAsyncThunk(
  'cluster/physicalBackupMaster',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.physicalBackupMaster(clusterName, serverId, baseURL)
      showSuccessBanner('Physical master backup successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Physical master backup failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const logicalBackup = createAsyncThunk('cluster/logicalBackup', async ({ clusterName, serverId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.logicalBackup(clusterName, serverId, baseURL)
    showSuccessBanner('Logical backup successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Logical backup failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const stopDatabase = createAsyncThunk('cluster/stopDatabase', async ({ clusterName, serverId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.stopDatabase(clusterName, serverId, baseURL)
    showSuccessBanner('Database is stopped!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Stopping database failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const startDatabase = createAsyncThunk('cluster/startDatabase', async ({ clusterName, serverId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.startDatabase(clusterName, serverId, baseURL)
    showSuccessBanner('Database has started!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    console.log('error in startDatabase::', error)
    showErrorBanner('Starting database failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const provisionDatabase = createAsyncThunk(
  'cluster/provisionDatabase',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.provisionDatabase(clusterName, serverId, baseURL)
      showSuccessBanner('Provision database successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Provision database failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const unprovisionDatabase = createAsyncThunk(
  'cluster/unprovisionDatabase',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.unprovisionDatabase(clusterName, serverId, baseURL)
      showSuccessBanner('Unprovision database successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Unprovision database failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const runRemoteJobs = createAsyncThunk('cluster/runRemoteJobs', async ({ clusterName, serverId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.runRemoteJobs(clusterName, serverId, baseURL)
    showSuccessBanner('Remote jobs started!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Remote jobs failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const optimizeServer = createAsyncThunk(
  'cluster/optimizeServer',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.optimizeServer(clusterName, serverId, baseURL)
      showSuccessBanner('Database optimize successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Database optimize failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const skipReplicationEvent = createAsyncThunk(
  'cluster/skipReplicationEvent',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.skipReplicationEvent(clusterName, serverId, baseURL)
      showSuccessBanner('Replication event skipped!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Skipping Replication event failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const toggleInnodbMonitor = createAsyncThunk(
  'cluster/toggleInnodbMonitor',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.toggleInnodbMonitor(clusterName, serverId, baseURL)
      showSuccessBanner('Toggle Innodb Monitor successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Toggle Innodb Monitor failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const toggleSlowQueryCapture = createAsyncThunk(
  'cluster/toggleSlowQueryCapture',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.toggleSlowQueryCapture(clusterName, serverId, baseURL)
      showSuccessBanner('Toggle Slow Query Capture successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Toggle Slow Query Capture failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const startSlave = createAsyncThunk('cluster/startSlave', async ({ clusterName, serverId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.startSlave(clusterName, serverId, baseURL)
    showSuccessBanner('Slave has started!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Starting slave failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const stopSlave = createAsyncThunk('cluster/stopSlave', async ({ clusterName, serverId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.stopSlave(clusterName, serverId, baseURL)
    showSuccessBanner('Slave has stopped!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Starting slave failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const toggleReadOnly = createAsyncThunk(
  'cluster/toggleReadOnly',
  async ({ clusterName, serverId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.toggleReadOnly(clusterName, serverId, baseURL)
      showSuccessBanner('Toggle readonly successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Toggle readonly failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const resetMaster = createAsyncThunk('cluster/resetMaster', async ({ clusterName, serverId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.resetMaster(clusterName, serverId, baseURL)
    showSuccessBanner('Reset Master successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Reset Master failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const resetSlaveAll = createAsyncThunk('cluster/resetSlaveAll', async ({ clusterName, serverId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.resetSlaveAll(clusterName, serverId, baseURL)
    showSuccessBanner('Reset Slave successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Reset Slave failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const cancelServerJob = createAsyncThunk(
  'cluster/cancelServerJob',
  async ({ clusterName, serverId, taskName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.cancelServerJob(clusterName, serverId, taskName, baseURL)
      showSuccessBanner(`Job ${taskName} cancelled successful!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Cancellation of job ${taskName} failed!`, error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const provisionProxy = createAsyncThunk('cluster/provisionProxy', async ({ clusterName, proxyId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.provisionProxy(clusterName, proxyId, baseURL)
    showSuccessBanner('Provision proxy successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Provision proxy failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const unprovisionProxy = createAsyncThunk(
  'cluster/unprovisionProxy',
  async ({ clusterName, proxyId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.unprovisionProxy(clusterName, proxyId, baseURL)
      showSuccessBanner('Unprovision proxy successful!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Unprovision proxy failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const startProxy = createAsyncThunk('cluster/startProxy', async ({ clusterName, proxyId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.startProxy(clusterName, proxyId, baseURL)
    showSuccessBanner('Starting proxy successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Starting proxy failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const stopProxy = createAsyncThunk('cluster/stopProxy', async ({ clusterName, proxyId }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.stopProxy(clusterName, proxyId, baseURL)
    showSuccessBanner('Stopping proxy successful!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Stopping proxy failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const runSysBench = createAsyncThunk('cluster/runSysBench', async ({ clusterName, thread }, thunkAPI) => {
  try {
    const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
    const { data, status } = await clusterService.runSysbench(clusterName, thread, baseURL)
    showSuccessBanner('Sysbench ran successfuly!', status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner('Sysbench failed!', error, thunkAPI)
    handleError(error, thunkAPI)
  }
})

export const runRegressionTests = createAsyncThunk(
  'cluster/runRegressionTests',
  async ({ clusterName, testName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.runRegressionTests(clusterName, testName, baseURL)
      showSuccessBanner('Regression test ran successfuly!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Regression test failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const getDatabaseService = createAsyncThunk(
  'cluster/getDatabaseService',
  async ({ clusterName, serviceName, dbId }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.getDatabaseService(clusterName, serviceName, dbId, baseURL)
      return { data, status }
    } catch (error) {
      handleError(error, thunkAPI)
    }
  }
)

export const updateLongQueryTime = createAsyncThunk(
  'cluster/updateLongQueryTime',
  async ({ clusterName, dbId, time }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.updateLongQueryTime(clusterName, dbId, time, baseURL)
      showSuccessBanner('Long query time updated!', status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner('Long query time update failed!', error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const checksumTable = createAsyncThunk(
  'cluster/checksumTable',
  async ({ clusterName, schema, table }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.checksumTable(clusterName, schema, table, baseURL)
      showSuccessBanner(`Checksum done for schema ${schema} and table ${table}!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Checksum failed for schema ${schema} and table ${table}!`, error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const toggleDatabaseActions = createAsyncThunk(
  'cluster/toggleDatabaseActions',
  async ({ clusterName, dbId, serviceName }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.toggleDatabaseActions(clusterName, serviceName, dbId, baseURL)
      showSuccessBanner(`Toggle ${serviceName} successful!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Toggle ${serviceName} failed!`, error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const addUser = createAsyncThunk(
  'cluster/addUser',
  async ({ clusterName, username, grants, roles }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.addUser(clusterName, username, grants, roles, baseURL)
      showSuccessBanner(`User is added successful!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Adding user failed!`, error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const updateGrants = createAsyncThunk(
  'cluster/updateGrants',
  async ({ clusterName, username, grants, roles }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.updateGrants(clusterName, username, grants, roles, baseURL)
      showSuccessBanner(`User is added successful!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Adding user failed!`, error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const dropUser = createAsyncThunk(
  'cluster/dropUser',
  async ({ clusterName, username, grants, roles }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.dropUser(clusterName, username, baseURL)
      showSuccessBanner(`User is added successful!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Adding user failed!`, error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const clusterSubscribe = createAsyncThunk('auth/clusterSubscribe', async ({  password, clusterName, baseURL }, thunkAPI) => {
  try {
    const { data, status } = await clusterService.clusterSubscribe(thunkAPI.getState().auth.user.username, password, clusterName, baseURL)
    showSuccessBanner(`Register user to peer cluster sent!`, status, thunkAPI)
    return { data, status }
  } catch (error) {
    showErrorBanner(`Register user to peer cluster failed!`, error, thunkAPI)
    const errorMessage = error.message || 'Request failed'
    const errorStatus = error.errorStatus || 500 // Default error status if not provided
    // Handle errors (including custom errorStatus)
    return thunkAPI.rejectWithValue({ errorMessage, errorStatus }) // Pass the entire Error object to the rejected action
  }
})

export const acceptSubscription = createAsyncThunk(
  'cluster/acceptSubscription',
  async ({ clusterName, username }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.acceptSubscription(clusterName, username, baseURL)
      showSuccessBanner(`Subscription accepted successfully!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Accept subscription failed!`, error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const rejectSubscription = createAsyncThunk(
  'cluster/rejectSubscription',
  async ({ clusterName, username }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.rejectSubscription(clusterName, username, baseURL)
      showSuccessBanner(`Subscription rejected successfully!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Reject subscription failed!`, error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

export const endSubscription = createAsyncThunk(
  'cluster/endSubscription',
  async ({ clusterName, username }, thunkAPI) => {
    try {
      const baseURL = thunkAPI.getState()?.auth?.baseURL || ''
      const { data, status } = await clusterService.endSubscription(clusterName, username, baseURL)
      showSuccessBanner(`Subscription ended successfully!`, status, thunkAPI)
      return { data, status }
    } catch (error) {
      showErrorBanner(`Failed to end subscription!`, error, thunkAPI)
      handleError(error, thunkAPI)
    }
  }
)

const initialState = {
  loading: false,
  error: null,
  clusterData: null,
  clusterAlerts: null,
  clusterMaster: null,
  clusterServers: null,
  clusterProxies: null,
  clusterCertificates: null,
  backupSnapshots: null,
  topProcess: null,
  jobs: null,
  shardSchema: null,
  queryRules: null,
  refreshInterval: 0,
  loadingStates: {
    switchOver: false,
    failOver: false,
    menuActions: false
  },
  database: {
    processList: null,
    status: {
      statusDelta: null,
      statusInnoDB: null
    },
    slowQueries: null,
    digestQueries: null,
    tables: null,
    errors: null,
    variables: null,
    serviceOpensvc: null,
    metadataLocks: null,
    responsetime: null
  }
}

export const clusterSlice = createSlice({
  name: 'cluster',
  initialState,
  reducers: {
    setRefreshInterval: (state, action) => {
      localStorage.setItem('refresh_interval', action.payload.interval)
      state.refreshInterval = action.payload.interval
    },
    pauseAutoReload: (state, action) => {
      if (action.payload.isPaused) {
        localStorage.setItem('pause_auto_reload', true)
      } else {
        localStorage.removeItem('pause_auto_reload')
      }
    },
    setCluster: (state, action) => {
      state.clusterData = action.payload.data
    },
    clearCluster: (state, action) => {
      Object.assign(state, initialState)
    }
  },
  extraReducers: (builder) => {
    builder.addMatcher(
      isAnyOf(
        getClusterData.fulfilled,
        getClusterAlerts.fulfilled,
        getClusterMaster.fulfilled,
        getClusterServers.fulfilled,
        getClusterProxies.fulfilled,
        getClusterCertificates.fulfilled,
        getDatabaseService.fulfilled,
        getTopProcess.fulfilled,
        getBackupSnapshot.fulfilled,
        getShardSchema.fulfilled,
        getQueryRules.fulfilled,
        getJobs.fulfilled
      ),
      (state, action) => {
        if (action.type.includes('getClusterData')) {
          state.clusterData = action.payload.data
        } else if (action.type.includes('getClusterAlerts')) {
          state.clusterAlerts = action.payload.data
        } else if (action.type.includes('getClusterMaster')) {
          state.clusterMaster = action.payload.data
        } else if (action.type.includes('getClusterServers')) {
          state.clusterServers = action.payload.data
        } else if (action.type.includes('getClusterProxies')) {
          state.clusterProxies = action.payload.data
        } else if (action.type.includes('getClusterCertificates')) {
          state.clusterCertificates = action.payload.data
        } else if (action.type.includes('getTopProcess')) {
          state.topProcess = action.payload.data
        } else if (action.type.includes('getBackupSnapshot')) {
          state.backupSnapshots = action.payload.data
        } else if (action.type.includes('getShardSchema')) {
          state.shardSchema = action.payload.data
        } else if (action.type.includes('getQueryRules')) {
          state.queryRules = action.payload.data
        } else if (action.type.includes('getJobs')) {
          state.jobs = action.payload.data
        } else if (action.type.includes('getDatabaseService')) {
          const { serviceName } = action.meta.arg
          if (serviceName === 'processlist') {
            state.database.processList = action.payload.data
          } else if (serviceName === 'slow-queries') {
            state.database.slowQueries = action.payload.data
          } else if (serviceName === 'digest-statements-pfs') {
            state.database.digestQueries = action.payload.data
          } else if (serviceName === 'tables') {
            state.database.tables = action.payload.data
          } else if (serviceName === 'status-delta') {
            state.database.status.statusDelta = action.payload.data
          } else if (serviceName === 'status-innodb') {
            state.database.status.statusInnoDB = action.payload.data
          } else if (serviceName === 'variables') {
            state.database.variables = action.payload.data
          } else if (serviceName === 'service-opensvc') {
            state.database.serviceOpensvc = action.payload.data
          } else if (serviceName === 'meta-data-locks') {
            state.database.metadataLocks = action.payload.data
          } else if (serviceName === 'query-response-time') {
            state.database.responsetime = action.payload.data
          }
        }
      }
    )

    builder.addMatcher(
      isAnyOf(
        switchOverCluster.pending,
        failOverCluster.pending,
        resetFailOverCounter.pending,
        resetSLA.pending,
        addServer.pending,
        toggleTraffic.pending,
        provisionCluster.pending,
        unProvisionCluster.pending,
        sendCredentials.pending,
        rotateDBCredential.pending,
        rollingOptimize.pending,
        rollingRestart.pending,
        rotateCertificates.pending,
        reloadCertificates.pending,
        cancelRollingRestart.pending,
        cancelRollingReprov.pending,
        bootstrapMasterSlave.pending,
        bootstrapMasterSlaveNoGtid.pending,
        bootstrapMultiMaster.pending,
        bootstrapMultiMasterRing.pending,
        bootstrapMultiTierSlave.pending,
        configReload.pending,
        configDiscoverDB.pending,
        configDynamic.pending,
        setMaintenanceMode.pending,
        promoteToLeader.pending,
        setAsUnrated.pending,
        setAsPreferred.pending,
        setAsIgnored.pending,
        reseedLogicalFromBackup.pending,
        reseedLogicalFromMaster.pending,
        reseedPhysicalFromBackup.pending,
        flushLogs.pending,
        physicalBackupMaster.pending,
        logicalBackup.pending,
        stopDatabase.pending,
        startDatabase.pending,
        provisionDatabase.pending,
        unprovisionDatabase.pending,
        runRemoteJobs.pending,
        optimizeServer.pending,
        skipReplicationEvent.pending,
        toggleInnodbMonitor.pending,
        toggleSlowQueryCapture.pending,
        startSlave.pending,
        stopSlave.pending,
        toggleReadOnly.pending,
        resetMaster.pending,
        resetSlaveAll.pending,
        provisionProxy.pending,
        unprovisionProxy.pending,
        startProxy.pending,
        stopProxy.pending
      ),
      (state, action) => {
        if (action.type.includes('switchOverCluster')) {
          state.loadingStates.switchOver = true
        } else if (action.type.includes('failOverCluster')) {
          state.loadingStates.failOver = true
        } else {
          state.loadingStates.menuActions = true
        }
      }
    )
    builder.addMatcher(
      isAnyOf(
        switchOverCluster.fulfilled,
        failOverCluster.fulfilled,
        resetFailOverCounter.fulfilled,
        resetSLA.fulfilled,
        addServer.fulfilled,
        toggleTraffic.fulfilled,
        provisionCluster.fulfilled,
        unProvisionCluster.fulfilled,
        sendCredentials.fulfilled,
        rotateDBCredential.fulfilled,
        rollingOptimize.fulfilled,
        rollingRestart.fulfilled,
        rotateCertificates.fulfilled,
        reloadCertificates.fulfilled,
        cancelRollingRestart.fulfilled,
        cancelRollingReprov.fulfilled,
        bootstrapMasterSlave.fulfilled,
        bootstrapMasterSlaveNoGtid.fulfilled,
        bootstrapMultiMaster.fulfilled,
        bootstrapMultiMasterRing.fulfilled,
        bootstrapMultiTierSlave.fulfilled,
        configReload.fulfilled,
        configDiscoverDB.fulfilled,
        configDynamic.fulfilled,
        setMaintenanceMode.fulfilled,
        promoteToLeader.fulfilled,
        setAsUnrated.fulfilled,
        setAsPreferred.fulfilled,
        setAsIgnored.fulfilled,
        reseedLogicalFromBackup.fulfilled,
        reseedLogicalFromMaster.fulfilled,
        reseedPhysicalFromBackup.fulfilled,
        flushLogs.fulfilled,
        physicalBackupMaster.fulfilled,
        logicalBackup.fulfilled,
        stopDatabase.fulfilled,
        startDatabase.fulfilled,
        provisionDatabase.fulfilled,
        unprovisionDatabase.fulfilled,
        runRemoteJobs.fulfilled,
        optimizeServer.fulfilled,
        skipReplicationEvent.fulfilled,
        toggleInnodbMonitor.fulfilled,
        toggleSlowQueryCapture.fulfilled,
        startSlave.fulfilled,
        stopSlave.fulfilled,
        toggleReadOnly.fulfilled,
        resetMaster.fulfilled,
        resetSlaveAll.fulfilled,
        provisionProxy.fulfilled,
        unprovisionProxy.fulfilled,
        startProxy.fulfilled,
        stopProxy.fulfilled
      ),
      (state, action) => {
        if (action.type.includes('switchOverCluster')) {
          state.loadingStates.switchOver = false
        } else if (action.type.includes('failOverCluster')) {
          state.loadingStates.failOver = false
        } else {
          state.loadingStates.menuActions = false
        }
      }
    )
    builder.addMatcher(
      isAnyOf(
        switchOverCluster.rejected,
        failOverCluster.rejected,
        resetFailOverCounter.rejected,
        resetSLA.rejected,
        addServer.rejected,
        toggleTraffic.rejected,
        provisionCluster.rejected,
        unProvisionCluster.rejected,
        sendCredentials.rejected,
        rotateDBCredential.rejected,
        rollingOptimize.rejected,
        rollingRestart.rejected,
        rotateCertificates.rejected,
        reloadCertificates.rejected,
        cancelRollingRestart.rejected,
        cancelRollingReprov.rejected,
        bootstrapMasterSlave.rejected,
        bootstrapMasterSlaveNoGtid.rejected,
        bootstrapMultiMaster.rejected,
        bootstrapMultiMasterRing.rejected,
        bootstrapMultiTierSlave.rejected,
        configReload.rejected,
        configDiscoverDB.rejected,
        configDynamic.rejected,
        setMaintenanceMode.rejected,
        promoteToLeader.rejected,
        setAsUnrated.rejected,
        setAsPreferred.rejected,
        setAsIgnored.rejected,
        reseedLogicalFromBackup.rejected,
        reseedLogicalFromMaster.rejected,
        reseedPhysicalFromBackup.rejected,
        flushLogs.rejected,
        physicalBackupMaster.rejected,
        logicalBackup.rejected,
        stopDatabase.rejected,
        startDatabase.rejected,
        provisionDatabase.rejected,
        unprovisionDatabase.rejected,
        runRemoteJobs.rejected,
        optimizeServer.rejected,
        skipReplicationEvent.rejected,
        toggleInnodbMonitor.rejected,
        toggleSlowQueryCapture.rejected,
        startSlave.rejected,
        stopSlave.rejected,
        toggleReadOnly.rejected,
        resetMaster.rejected,
        resetSlaveAll.rejected,
        provisionProxy.rejected,
        unprovisionProxy.rejected,
        startProxy.rejected,
        stopProxy.rejected
      ),
      (state, action) => {
        if (action.type.includes('switchOverCluster')) {
          state.loadingStates.switchOver = false
        } else if (action.type.includes('failOverCluster')) {
          state.loadingStates.failOver = false
        } else {
          state.loadingStates.menuActions = false
        }
      }
    )
  }
})

export const { setRefreshInterval, setCluster, clearCluster, pauseAutoReload } = clusterSlice.actions

// this is for configureStore
export default clusterSlice.reducer
