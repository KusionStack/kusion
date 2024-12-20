import React, { useEffect, useState } from 'react'
import { Button, Card, Col, Input, message, Popconfirm, Row, Table, Tabs, Tooltip } from 'antd'
import {
  PlusOutlined,
  SortDescendingOutlined,
  SortAscendingOutlined,
  SearchOutlined,
} from '@ant-design/icons'
import { WorkspaceService } from '@kusionstack/kusion-api-client-sdk';

import styles from './styles.module.less'
import BackWithTitle from '@/components/backWithTitle';
import { useLocation, useNavigate, useSearchParams, useParams } from 'react-router-dom';
import YamlEditor from '@/components/yamlEditor';
import { mockYaml } from '@/utils/tools';
import EditYamlDrawer from '../components/editYamlDrawer';
import MarkdownDrawer from '../components/markdownDrawer';

const orderIconStyle: React.CSSProperties = {
  marginLeft: 0,
}

const WorkspaceDetail = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [urlSearchParams] = useSearchParams();
  const urlParams = useParams();
  const [open, setOpen] = useState(false)
  const [openMod, setOpenMod] = useState(false)
  const [activeKey, setActiveKey] = useState('yaml');
  const [yamlData, setYamlData] = useState<any>();
  const [workspaceModules,setWorkspaceModules] = useState([]);

  console.log(location, urlSearchParams, urlParams, "====location====")

  async function getWorkspaceConfigs(workspaceId) {
    console.log(Number(workspaceId), "===workspaceId===")
    const response: any = await WorkspaceService.getWorkspaceConfigs({
      path: {
        id: workspaceId
      }
    })
    if (response?.data?.success) {
      const tempData = response?.data?.data?.modules;
      const list = Object.keys(tempData)?.map(key => {
        return {
          moduleName: key,
          ...tempData?.[key]
        }
      })
      setWorkspaceModules(list)
      setYamlData(JSON.stringify(response?.data?.data || {}, null, 2))
    }
    console.log(response?.data, "=========getWorkspaceConfigs=======")
  }

  async function getWorkspaceConfigsModules(workspaceId) {
    console.log(Number(workspaceId), "===workspaceId===")
    const response: any = await WorkspaceService.getWorkspaceConfigs({
      path: {
        id: workspaceId
      }
    })
    if (response?.data?.success) {
      setYamlData(JSON.stringify(response?.data?.data || {}, null, 2))
    }
    console.log(response?.data, "=========getWorkspaceConfigs=======")
  }


  useEffect(() => {
    if (urlParams?.workspaceId) {
      getWorkspaceConfigs(urlParams?.workspaceId)
    }
  }, [urlParams?.workspaceId])

  function handleAdd() {
    setOpen(true)
  }



  function handleSubmit(values) {
    console.log(values, "=====handleSubmit values=====")
  }
  function handleClose() {
    setOpen(false)
  }

  function validateYaml() {
    setOpen(false)
  }


  function handleBack() {
    navigate("/workspaces")
  }

  const items = [
    {
      key: 'yaml',
      label: 'workspace.yaml',
    },
    {
      key: 'modules',
      label: '可用 modules',
    },
  ]

  function handleTabsChange(key) {
    setActiveKey(key)
  }

  function handleEdit() {
    setOpen(true)
  }

  function generateMod() {
    setOpenMod(true)
  }

  const columns = [
    {
      key: 'moduleName',
      title: "Name",
      dataIndex: "moduleName"
    },
    {
      key: 'Registry',
      title: 'Registry',
      dataIndex: "path",
    }
  ]

  console.log(workspaceModules, "===workspaceModules===")

  return (

    <div className={styles.workspace_detail_container}>
      <BackWithTitle title="Workspaces" handleBack={handleBack} />
      <Card>
        <div className={styles.workspace_detail}>
          <Tabs activeKey={activeKey} items={items} onChange={handleTabsChange} />
          {
            activeKey === 'yaml' && <>
              <Button type='primary' style={{ marginBottom: 15 }} onClick={handleEdit}>Edit Yaml</Button>
              <YamlEditor readOnly={true} value={yamlData} themeMode={'DARK'} />
              <EditYamlDrawer yamlData={yamlData} open={open} handleClose={handleClose} handleSubmit={handleSubmit} validateYaml={validateYaml} />
            </>
          }
          {
            activeKey === 'modules' && <>
              <Button type='primary' style={{ marginBottom: 15 }} onClick={generateMod}>Generate kcl.mod</Button>
              <Table columns={columns} dataSource={workspaceModules} />
            </>
          }
        </div>
      </Card>
    </div>

  )
}

export default WorkspaceDetail
