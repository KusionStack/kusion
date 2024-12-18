import React, { useEffect, useRef, useState, memo } from 'react'
import { renderToString } from "react-dom/server";
import G6 from '@antv/g6';
import type { IAbstractGraph, IG6GraphEvent } from '@antv/g6'
import insertCss from "insert-css"
import { registerFlowLine, registerResourceNode, getEdgesLayer } from "./register";
import styles from './style.module.less'
import { Tag } from 'antd';
import { generateG6GraphData } from '@/utils/tools';

insertCss(`
  .g6-component-tooltip {
    background-color: #f0f5ff;
    padding: 10px 30px;
    box-shadow: rgb(174, 174, 174) 0px 0px 10px;
    border-top: 2px solid #2f54eb;
    color: #646566;
  }
  .tooltip-item {
    margin-bottom: 10px;
  }
  .type {
    background: rgba(255, 0, 0, .5);
    padding: 2px 5px;
    border-radius: 6px;
    color: #fff;
  }
`);

const tooltip = new G6.Tooltip({
  offsetX: 10,
  offsetY: 10,
  // the types of items that allow the tooltip show up
  // 允许出现 tooltip 的 item 类型
  itemTypes: ['node', 'edge'],
  // custom the tooltip's content
  // 自定义 tooltip 内容
  getContent: (e) => {
    const { nodeData, label, id }: any = e.item.getModel();
    const typeList = nodeData?.resourceType?.split('/');
    const type = typeList?.[typeList?.length - 1]
    const outDiv = document.createElement('div');
    outDiv.style.width = 'fit-content';
    // outDiv.style.padding = '0px 0px 10px 0px';
    outDiv.innerHTML = `
      <h4>${label || id}</h4>
      <div>
        <div class="tooltip-item">Name: ${label || id}</div>
        <div class="tooltip-item">Type: <span class="type">${type}</span></div>
        <div class="tooltip-item">Status: <span class="type">${nodeData?.status}</span></div>
        <div class="tooltip-item">cloudResourceID: ${nodeData?.cloudResourceID}</div>
        <div class="tooltip-item">iamResourceID: ${nodeData?.iamResourceID}</div>
      </div>`;
    return outDiv;
  },
});

const OverviewTooltip = memo((props: any) => {
  const model = props?.hiddenButtonInfo?.e.item?.get('model')
  const boxStyle: any = {
    background: '#fff',
    border: '1px solid #f5f5f5',
    position: 'absolute',
    top: props?.hiddenButtonInfo?.y - 60 || -500,
    left: props?.hiddenButtonInfo?.x + 60 || -500,
    zIndex: 5,
    padding: 10,
    borderRadius: 8,
    fontSize: 12,
    borderTop: '5px solid #1677ff'
  }
  const itemStyle = {
    color: '#646566',
    margin: '10px 5px',
  }
  return (
    <div style={boxStyle}>
      <div style={itemStyle}>
        {model?.label}
      </div>
      <div style={itemStyle}>
        {model?.kind}
      </div>
    </div>
  )
})


const G6Tree = ({ graphData }) => {

  const graphRef = useRef<any>()

  const [hiddenButtontooltip, setHiddenButtontooltip] = useState<{
    x: number
    y: number
    e?: IG6GraphEvent
  }>({ x: -500, y: -500, e: undefined })
  const [tooltipopen, setTooltipopen] = useState(false)
  const [itemWidth, setItemWidth] = useState<number>(100)

  function register() {
    registerResourceNode();
    registerFlowLine();
  }

  function handleMouseEnter(evt) {
    const model = evt?.item?.get('model')
    // graph.setItemState(evt.item, 'hoverState', true)
    const { x, y } = graphRef.current?.getCanvasByPoint(model.x, model.y)
    const node = graphRef.current?.findById(model.id)?.getBBox()
    if (node) {
      setItemWidth(node?.maxX - node?.minX)
    }
    setTooltipopen(true)
    setHiddenButtontooltip({ x, y, e: evt })
  }
  function handleMouseLeave(evt) {
    // graph.setItemState(evt.item, 'hoverState', false)
    setTooltipopen(false)
  }

  function updateSize() {
    const container = document.getElementById('mountNode');
    if (container === null) {
      return;
    }
    const width = container.scrollWidth || window.outerWidth - 90;
    const height = container.scrollHeight || window.outerHeight - 150;
    if (graphRef.current) {
      graphRef.current.changeSize(width, height);
    }
  }

  function initData() {
    // const data = {
    //   // 点集
    //   nodes: [
    //     {
    //       id: '1',
    //       label: 'zk',
    //       kind: 'Helm',
    //       health: 'Healthy',
    //       status: 'Synced',
    //       logoIcon: {
    //         img: 'https://blazehu.com/images/k8s/status/helm.svg',
    //         offsetY: 0,
    //       },
    //     },
    //     {
    //       id: '2',
    //       label: 'zk-zookeeper',
    //       kind: 'Service',
    //       health: 'Healthy',
    //       status: 'Synced',
    //     },
    //     {
    //       id: '3',
    //       label: 'zk-zookeeper-headless',
    //       kind: 'Service',
    //       health: 'Healthy',
    //       status: 'Synced',
    //     },
    //     {
    //       id: '4',
    //       label: 'zk-zookeeper',
    //       kind: 'StatefulSet',
    //       health: 'Progressing',
    //       status: 'Synced',
    //       phase: 'StartSync',
    //     },
    //     {
    //       id: '5',
    //       label: 'zk-zookeeper',
    //       kind: 'Endpoints',
    //     },
    //     {
    //       id: '6',
    //       label: 'zk-zookeeper-k2vpn',
    //       kind: 'EndpointSlice',
    //     },
    //     {
    //       id: '7',
    //       label: 'zk-zookeeper-headless',
    //       kind: 'Endpoints',
    //     },
    //     {
    //       id: '8',
    //       label: 'zk-zookeeper-headless-7bz9d',
    //       kind: 'EndpointSlice',
    //     },
    //     {
    //       id: '9',
    //       label: 'zk-zookeeper-headless-8mbqm',
    //       kind: 'EndpointSlice',
    //     },
    //     {
    //       id: '10',
    //       label: 'zk-zookeeper-0',
    //       kind: 'PersistentVolumeClaim',
    //       health: 'Healthy',
    //     },
    //     {
    //       id: '11',
    //       label: 'zk-zookeeper-0',
    //       kind: 'Pod',
    //       health: 'Progressing',
    //     },
    //     {
    //       id: '12',
    //       label: 'zk-zookeeper-565fd5c886',
    //       kind: 'ControllerRevision',
    //     }
    //   ],
    //   // 边集
    //   edges: [
    //     {
    //       source: '1', // String，必须，起始点 id
    //       target: '2', // String，必须，目标点 id
    //     },
    //     {
    //       source: '1',
    //       target: '3',
    //     },
    //     {
    //       source: '1',
    //       target: '4',
    //     },
    //     {
    //       source: '2',
    //       target: '5',
    //     },
    //     {
    //       source: '2',
    //       target: '6',
    //     },
    //     {
    //       source: '3',
    //       target: '7',
    //     },
    //     {
    //       source: '3',
    //       target: '8',
    //     },
    //     {
    //       source: '3',
    //       target: '9',
    //     },
    //     {
    //       source: '4',
    //       target: '10',
    //     },
    //     {
    //       source: '4',
    //       target: '11',
    //     },
    //     {
    //       source: '4',
    //       target: '12',
    //     },
    //   ],
    // };
    const data = generateG6GraphData(graphData)
    console.log(data, "====data====")
    const edgesLayer = getEdgesLayer(data.edges || []);
    const valList: any = Object.values(edgesLayer);
    const maxLayerCount = Math.max(...valList);
    return { maxLayerCount, data };
  }

  function createTree() {
    if (graphRef.current) {
      graphRef.current.destroy();
    }
    if (!graphRef.current) {
      const container = document.getElementById('mountNode');
      if (container === null) {
        return;
      }
      const width = container.scrollWidth || window.outerWidth - 90;
      const height = container.scrollHeight || window.outerHeight - 150;
      const defaultEdgeStyle = {
        stroke: '#e2e2e2',
        // stroke: 'red',
        endArrow: {
          path: 'M 0,0 L 8,4 L 8,-4 Z',
          fill: '#e2e2e2',
          stroke: '#bae7ff',

          // fill: '#2f54eb',
          // opacity: 0.4,
          // stroke: '#2f54eb',
          lineWidth: 2,
        },
      };
      const defaultNodeStyle = {
        stroke: '#00000000',
        // shadowColor: '#1677ff',
        // shadowColor: '#ccd6dd',
        // shadowOffsetX: 1,
        // shadowOffsetY: 1,
        // shadowBlur: 1,
        // fill: '#f0f5ff',
        fill: '#fff',
        cursor: 'pointer',
      };
      // const defaultNodeSize = [345, 72]; // [width, height]
      const defaultNodeSize = [300, 72]; // [width, height]
      const defaultLogo = {
        width: 32,
        height: 32,
      };
      const graphTmp = new G6.Graph({
        container: 'mountNode',
        width,
        height,
        fitView: true,
        // fitCenter: true,
        modes: {
          default: [{
            type: 'drag-canvas',
            // ... 其他配置
          }, {
            type: 'scroll-canvas',
            direction: 'y',
            scalableRange: height * -0.5,
            // ... 其他配置
          }],
        },
        plugins: [tooltip],
        layout: {
          type: 'dagre',
          rankdir: 'LR',
          align: 'DL',
          nodesep: 30, // 可选
          // ranksep: 60, // 可选
          nodesepFunc: () => 1,
          ranksepFunc: () => 1,
        },
        defaultNode: {
          type: 'resource',
          size: defaultNodeSize,
          style: defaultNodeStyle,
          logoIcon: defaultLogo,
          stateIcon: {
            show: false,
          },
          preRect: {
            show: false,
          },
        },
        defaultEdge: {
          type: 'flow-line', // line、flow-line、circle-running
          style: defaultEdgeStyle,
          size: 1,
          color: '#e2e2e2',
        },
        edgeStateStyles: {
          hover: {
            lineWidth: 6,
          },
        },
      });
      return graphTmp;
    }
  }

  function initTree() {
    const { data } = initData();
    graphRef.current = createTree()
    updateSize();
    graphRef.current.read(data);
    graphRef.current.zoomTo(0.75, { x: 128, y: 369 }, true, { duration: 10 });
    graphRef.current.on('node:click', (evt) => {
      console.log(evt?.item?.get('model'), "NODE click")
      // graph.setItemState(evt.item, 'hover', true)
    })
    // graphRef.current.on('node:mouseenter', evt => {
    //   console.log("NODE mouseenter")
    //   handleMouseEnter(evt)
    //   // graph.setItemState(evt.item, 'hover', false)
    // })
    // graphRef.current.on('node:mouseenter', evt => {
    //   console.log("NODE mouseenter")
    //   // handleMouseLeave(evt)
    //   // graph.setItemState(evt.item, 'hover', false)
    // })
  }


  useEffect(() => {
    register()
    initTree()
    return () => {
      try {
        if (graphRef.current) {
          graphRef.current.destroy()
          graphRef.current = null
        }
      } catch (error) { }
    }
  }, [graphData])


  return (
    <div className={styles.kusion_g6_tree}>
      <div id="mountNode"></div>
      {tooltipopen ? (
        <OverviewTooltip
          itemWidth={itemWidth}
          hiddenButtonInfo={hiddenButtontooltip}
          open={tooltipopen}
        />
      ) : null}
    </div>
  )
}

export default G6Tree
