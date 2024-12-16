import { createColumnHelper } from '@tanstack/react-table'
import React, { useEffect, useMemo, useState } from 'react'
import { DataTable } from '../../components/DataTable'
import AccordionComponent from '../../components/AccordionComponent'
import styles from './styles.module.scss'
import UserGrantModal from '../../components/Modals/UserGrantModal'
import RMButton from '../../components/RMButton'
import RMIconButton from '../../components/RMIconButton'
import { HiUserGroup } from 'react-icons/hi'

function Users({ selectedCluster, user }) {
  const [data, setData] = useState([])
  const [selectedUser, setSelectedUser] = useState(null)
  const [isUserGrantModalOpen, setIsUserGrantModalOpen] = useState(null)
  const columnHelper = createColumnHelper()

  useEffect(() => {
    if (selectedCluster?.apiUsers) {
      const result = Object.entries(selectedCluster?.apiUsers).map(([key, value]) => ({
        user: key,
        ...value
      }));

      setData(result)
    }
  }, [selectedCluster?.apiUsers])

  const openUserGrantModal = () => {
    setIsUserGrantModalOpen(true)
  }

  const closeUserGrantModal = () => {
    setIsUserGrantModalOpen(false)
  }

  const columns = useMemo(
    () => [
      columnHelper.accessor((row) => row.user, {
        cell: (info) => info.getValue(),
        header: 'User',
        id: 'user'
      }),
      columnHelper.accessor((row) => row.isExternal, {
        cell: (info) => info.getValue(),
        header: 'Is Git Validated',
        id: 'isExternal'
      }),
      columnHelper.accessor((row) => {
        return Object.entries(row.roles).filter(([_, v]) => v).map(([role, _]) => (<span>{role}</span>)).reduce((r, n) => {
          if (r.length) {
            r.push(<br />)
          }
          r.push(n)
          return r
        }, [])
      }, {
        cell: (info) => info.getValue(),
        header: 'Roles',
        id: 'roles'
      }),
      columnHelper.accessor((row) => (<RMIconButton icon={HiUserGroup} onClick={(e) => { e.stopPropagation(); setSelectedUser(row); openUserGrantModal()}} />), {
        cell: (info) => info.getValue(),
        header: 'Grants',
        id: 'grants'
      })
    ],
    []
  )

  
  return (
    <>
    <AccordionComponent
      heading={'USERS'}
      allowToggle={false}
      className={styles.accordion}
      panelSX={{ overflowX: 'auto', p: 0 }}
      body={<DataTable data={data} columns={columns} className={styles.table} />}
    />
    {isUserGrantModalOpen && <UserGrantModal clusterName={selectedCluster.name} user={selectedUser} isOpen={isUserGrantModalOpen} closeModal={closeUserGrantModal} />}
    </>
  )
}

export default Users
