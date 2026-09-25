import { Layout } from '@douyinfe/semi-ui';
import { Outlet } from 'react-router-dom'
import styles from './App.module.css'
import { Menu } from '@/components/menu';
import { HeaderBar } from '@/components/headerBar';
const { Header, Sider, Content } = Layout;

function App() {
  return (
    <Layout className={styles.root}>
      <Sider className={styles.root_sider_bar}>
        <Menu />
      </Sider>
      <Layout className={styles.root_right}>
        <Header className={styles.root_header}>
          <HeaderBar />
        </Header>
        <Content
          className={styles.root_content}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}

export default App
