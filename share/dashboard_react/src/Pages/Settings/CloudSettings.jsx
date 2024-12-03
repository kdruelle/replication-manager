import { Flex } from '@chakra-ui/react'
import React , { useState, useEffect } from 'react'
import styles from './styles.module.scss'
import RMSwitch from '../../components/RMSwitch'
import { useDispatch ,useSelector } from 'react-redux'
import TableType2 from '../../components/TableType2'
import Dropdown from '../../components/Dropdown'
import { convertObjectToArrayForDropdown, formatBytes } from '../../utility/common'
import { setSetting, switchSetting } from '../../redux/settingsSlice'

function CloudSettings({ selectedCluster, user }) {
  const dispatch = useDispatch()

  const {
    globalClusters: { monitor }
  } = useSelector((state) => state)

  const [planOptions, setPlanOptions] = useState([])
  const getPlanOptions = (plist = []) => [{ name: "No Plan", value: '' } ,...plist?.map((obj) => ({ name: obj.plan, value: obj.plan }))]

  const onPlanChange = (option) => {
    setPlan(option.value)
    //setPlanDetails(planOptions.find((obj) => obj.plan === option.value))
    //console.log(plans,option)
  }
  useEffect(() => {
     if (monitor?.servicePlans) {
       setPlanOptions(getPlanOptions(monitor.servicePlans))
     }
  }, [monitor?.servicePlans])

  const dataObject = [
    ...(selectedCluster?.config?.cloud18
      ? [
          {
            key: 'For Sale',
            value: (
              <RMSwitch
                confirmTitle={'Confirm switch settings for cloud18-shared?'}
                onChange={() =>
                  dispatch(switchSetting({ clusterName: selectedCluster?.name, setting: 'cloud18-shared' }))
                }
                isDisabled={user?.grants['cluster-settings'] == false}
                isChecked={selectedCluster?.config?.cloud18Shared}
              />
            )
          },
          {
            key: 'Cluster Plan',
            value: (
              <Flex className={styles.dropdownContainer}>
                <Dropdown
                  options={planOptions}
                  id='plan'
                  className={styles.dropdownButton}
                  selectedValue={selectedCluster?.config?.provServicePlan}
                  confirmTitle={`Confirm plan change to`}
                  onChange={(option) => {
                    dispatch(
                      setSetting({
                        clusterName: selectedCluster?.name,
                        setting: 'prov-service-plan',
                        value: option
                      })
                    )
                  }}
                />
              </Flex>
        )
       }

        ]
      : [])
  ]

  return (
    <Flex justify='space-between' gap='0'>
      <TableType2 dataArray={dataObject} className={styles.table} />
    </Flex>
  )
}

export default CloudSettings
