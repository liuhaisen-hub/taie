import { Layout } from '@douyinfe/semi-ui';
import { Outlet } from 'react-router-dom'
import styles from './App.module.css'
import { HeaderBar } from '@/components/headerBar';
import { SiderBard } from './components/siderbar';
import { LayoutProvider } from './hooks/layout';
const { Header, Sider, Content } = Layout;

function App() {
  return (
    <LayoutProvider>
      <Layout className={styles.root}>
        <Sider className={styles.root_sider_bar}>
          <SiderBard />
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
    </LayoutProvider>

  )
}

export default App
