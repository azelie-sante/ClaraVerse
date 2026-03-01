import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'intro',
    {
      type: 'category',
      label: 'Getting Started',
      items: [
        'getting-started/installation',
        'getting-started/configuration',
        'getting-started/quickstart',
        'getting-started/updating',
      ],
      collapsed: false,
    },
    {
      type: 'category',
      label: 'Features',
      items: [
        'features/chat',
        'features/workflows',
        'features/nexus',
        'features/routines',
        'features/channels',
        'features/integrations',
        'features/artifacts',
        'features/memory',
        'features/devices',
        'features/companion',
        'features/ai-docs',
        'features/local-ai',
      ],
    },
    {
      type: 'category',
      label: 'Providers',
      items: [
        'providers/overview',
        'providers/openai',
        'providers/anthropic',
        'providers/google',
        'providers/ollama',
        'providers/custom',
      ],
    },
    {
      type: 'category',
      label: 'Self-Hosting',
      items: [
        'self-hosting/architecture',
        'self-hosting/docker',
        'self-hosting/environment-variables',
        'self-hosting/security',
        'self-hosting/troubleshooting',
      ],
    },
    {
      type: 'category',
      label: 'Admin Guide',
      items: [
        'admin/overview',
        'admin/user-management',
        'admin/provider-management',
        'admin/model-management',
        'admin/analytics',
      ],
    },
    'faq',
  ],
};

export default sidebars;
