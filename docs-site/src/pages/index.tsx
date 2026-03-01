import type {ReactNode} from 'react';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import useBaseUrl from '@docusaurus/useBaseUrl';
import Layout from '@theme/Layout';
import styles from './index.module.css';

const features = [
  {
    title: 'AI Chat + Skills',
    icon: '\u{1F4AC}',
    desc: 'Context-aware tools activate mid-conversation \u2014 web search, image gen, data analysis.',
    link: '/docs/features/chat',
  },
  {
    title: 'Nexus',
    icon: '\u{1F9E0}',
    desc: 'Assign long-running tasks like research and coding. Track on a Kanban board.',
    link: '/docs/features/nexus',
  },
  {
    title: 'Workflows',
    icon: '\u26A1',
    desc: 'Drag-and-drop visual builder with parallel execution and scheduling.',
    link: '/docs/features/workflows',
  },
  {
    title: 'Routines',
    icon: '\u{1F504}',
    desc: 'Scheduled automations that run and report back via Telegram.',
    link: '/docs/features/routines',
  },
  {
    title: 'Channels',
    icon: '\u{1F4F1}',
    desc: 'Talk to Clara from Telegram. Get reports delivered to your phone.',
    link: '/docs/features/channels',
  },
  {
    title: '150+ Integrations',
    icon: '\u{1F517}',
    desc: 'Slack, GitHub, Jira, Notion, and more \u2014 built in, no MCP required.',
    link: '/docs/features/integrations',
  },
  {
    title: 'Devices',
    icon: '\u{1F5A5}\uFE0F',
    desc: 'Connect all your machines. Clara reaches MCP on any device remotely.',
    link: '/docs/features/devices',
  },
  {
    title: 'Local AI',
    icon: '\u{1F3E0}',
    desc: 'Ollama & LM Studio auto-detected. Run models locally, zero config.',
    link: '/docs/features/local-ai',
  },
];

function FeatureCard({title, icon, desc, link}: {
  title: string; icon: string; desc: string; link: string;
}): ReactNode {
  return (
    <Link to={link} className={styles.featureLink}>
      <div className="glass-card">
        <span className={styles.featureIcon}>{icon}</span>
        <h3 className={styles.featureTitle}>{title}</h3>
        <p className={styles.featureDesc}>{desc}</p>
      </div>
    </Link>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  const bannerUrl = useBaseUrl('/img/image-banner.png');
  return (
    <Layout title="Home" description={siteConfig.tagline}>
      {/* Hero */}
      <header className={styles.hero}>
        <div className={styles.heroGlow} />
        <div className={styles.heroContent}>
          <img src={bannerUrl} alt="ClaraVerse" className={styles.banner} />
          <h1 className={styles.title}>{siteConfig.tagline}</h1>
          <p className={styles.subtitle}>
            Built by the community, for the community.<br />
            Private AI that respects your freedom.
          </p>
          <div className={styles.cta}>
            <Link className="btn-primary" to="/docs/getting-started/installation">
              Get Started
            </Link>
            <Link className="btn-secondary" to="https://github.com/claraverse/ClaraVerse-Scarlet-OSS">
              GitHub
            </Link>
          </div>
        </div>
      </header>

      {/* Features */}
      <section className={styles.features}>
        <h2 className={styles.sectionTitle}>Everything you need</h2>
        <p className={styles.sectionSub}>A complete AI workspace \u2014 self-hosted, private, extensible.</p>
        <div className={styles.grid}>
          {features.map((f) => <FeatureCard key={f.title} {...f} />)}
        </div>
      </section>

      {/* Quick info strip */}
      <section className={styles.strip}>
        <div className={styles.stripInner}>
          <div className={styles.stat}>
            <span className={styles.statVal}>150+</span>
            <span className={styles.statLabel}>Integrations</span>
          </div>
          <div className={styles.stat}>
            <span className={styles.statVal}>BYOK</span>
            <span className={styles.statLabel}>Bring Your Keys</span>
          </div>
          <div className={styles.stat}>
            <span className={styles.statVal}>100%</span>
            <span className={styles.statLabel}>Self-Hosted</span>
          </div>
          <div className={styles.stat}>
            <span className={styles.statVal}>AGPL-3.0</span>
            <span className={styles.statLabel}>Open Source</span>
          </div>
        </div>
      </section>
    </Layout>
  );
}
