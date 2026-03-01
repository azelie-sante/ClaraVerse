import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'ClaraVerse',
  tagline: 'Your Private AI Workspace',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://claraverse-docs.netlify.app',
  baseUrl: '/',

  organizationName: 'claraverse',
  projectName: 'ClaraVerse-Scarlet-OSS',

  onBrokenLinks: 'throw',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl:
            'https://github.com/claraverse/ClaraVerse-Scarlet-OSS/tree/main/docs-site/',
        },
        blog: {
          showReadingTime: true,
          blogSidebarTitle: 'All Posts',
          blogSidebarCount: 'ALL',
          postsPerPage: 10,
          blogTitle: 'ClaraVerse Blog',
          blogDescription:
            'Stories, technical deep-dives, and updates from the ClaraVerse team -- building the open-source AI workspace.',
          feedOptions: {
            type: ['rss', 'atom'],
            title: 'ClaraVerse Blog',
          },
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/clara-mascot.png',
    colorMode: {
      defaultMode: 'dark',
      disableSwitch: true,
      respectPrefersColorScheme: false,
    },
    navbar: {
      title: 'ClaraVerse',
      logo: {
        alt: 'ClaraVerse Logo',
        src: 'img/logo.png',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
        },
        {
          to: '/blog',
          label: 'Blog',
          position: 'left',
        },
        {
          href: 'https://github.com/claraverse/ClaraVerse-Scarlet-OSS',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Getting Started',
              to: '/docs/getting-started/installation',
            },
            {label: 'Features', to: '/docs/features/chat'},
            {label: 'Self-Hosting', to: '/docs/self-hosting/architecture'},
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/claraverse/ClaraVerse-Scarlet-OSS',
            },
            {
              label: 'Discord',
              href: 'https://discord.gg/claraverse',
            },
          ],
        },
      ],
      copyright: `Copyright \u00a9 ${new Date().getFullYear()} ClaraVerse. AGPL-3.0 License.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'yaml', 'json', 'go', 'docker'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
