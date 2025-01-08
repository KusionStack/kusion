import React, { useEffect, useState } from 'react'
import { Button, Card, message, Table, Tabs } from 'antd'
import { WorkspaceService } from '@kusionstack/kusion-api-client-sdk';
import BackWithTitle from '@/components/backWithTitle';
import { useNavigate, useParams } from 'react-router-dom';
import YamlEditor from '@/components/yamlEditor';
import { josn2yaml, yaml2json } from '@/utils/tools'
import EditYamlDrawer from '../components/editYamlDrawer';
import MarkdownDrawer from '../components/markdownDrawer';

import styles from './styles.module.less'
import { DEFAULT_WORKSPACE_YAML } from '@/utils/constants';

const WorkspaceDetail = () => {
  const navigate = useNavigate();
  const urlParams = useParams();
  const [open, setOpen] = useState(false)
  const [openMod, setOpenMod] = useState(false)
  const [activeKey, setActiveKey] = useState('yaml');
  const [yamlData, setYamlData] = useState<any>();
  const [workspaceModules, setWorkspaceModules] = useState([]);
  const [markdown, setMarkdown] = useState('')

  async function getWorkspaceConfigs(workspaceId) {
    const response: any = await WorkspaceService.getWorkspaceConfigs({
      path: {
        workspaceID: workspaceId
      }
    })
    if (response?.data?.success) {
      console.log(response?.data?.data, "====response?.data?.data===")
      const tempData = response?.data?.data?.modules;
      const list = tempData && Object.keys(tempData)?.map(key => {
        return {
          moduleName: key,
          ...tempData?.[key]
        }
      })
      setWorkspaceModules(list)
      const yamlStr = Object?.keys(response?.data?.data)?.length > 0 ? JSON.stringify(response?.data?.data || {}, null, 2) : JSON.stringify(DEFAULT_WORKSPACE_YAML, null, 2)
      console.log(yamlStr, "====yamlStr====")
      setYamlData(josn2yaml(yamlStr)?.data)
    } else {
      message.error(response?.data?.message)
    }
  }

  useEffect(() => {
    if (urlParams?.workspaceId) {
      getWorkspaceConfigs(urlParams?.workspaceId)
    }
  }, [urlParams?.workspaceId])



  async function handleSubmit(yamlStr) {
    const response: any = await WorkspaceService.updateWorkspaceConfigs({
      body: yamlStr ? yaml2json(yamlStr)?.data : {},
      path: {
        workspaceID: Number(urlParams?.workspaceId)
      },
    })
    if (response?.data?.success) {
      message.success('Update Successful')
      getWorkspaceConfigs(urlParams?.workspaceId)
      setOpen(false)
    } else {
      message.error(response?.data?.message)
    }
  }
  function handleClose() {
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

  async function generateMod() {
    const response: any = await WorkspaceService.createWorkspaceModDeps({
      path: {
        workspaceID: Number(urlParams?.workspaceId)
      }
    })
    if (response?.data?.success) {
      setOpenMod(true)
      setMarkdown(response?.data?.data)
    } else {
      message.error('Generate Failed')
    }


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
    },
    {
      key: 'version',
      title: 'Version',
      dataIndex: "version",
    },
  ]

  function handleModClose() {
    setOpenMod(false)
  }

  return (

    <div className={styles.workspace_detail_container}>
      <BackWithTitle title="Workspaces" handleBack={handleBack} />
      <Card>
        <div className={styles.workspace_detail}>
          <Tabs activeKey={activeKey} items={items} onChange={handleTabsChange} />
          {
            activeKey === 'yaml' && <>
              <Button type='primary' style={{ marginBottom: 15 }} onClick={handleEdit}>Edit Yaml</Button>
              {
                yamlData && <YamlEditor readOnly={true} value={yamlData} themeMode={'DARK'} />
              }
              {
                open && yamlData && <EditYamlDrawer yamlData={yamlData} open={open} handleClose={handleClose} handleSubmit={handleSubmit} />
              }
            </>
          }
          {
            activeKey === 'modules' && <>
              <Button type='primary' style={{ marginBottom: 15 }} onClick={generateMod}>Generate kcl.mod</Button>
              <Table columns={columns} dataSource={workspaceModules} />
              <MarkdownDrawer open={openMod} handleClose={handleModClose} markdown={markdown} />
            </>
          }
        </div>
      </Card>
    </div>

  )
}

export default WorkspaceDetail
