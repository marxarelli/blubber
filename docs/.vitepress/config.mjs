import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Blubber",
  description: "Blubber Documentation",
  base: '/releng/blubber/',
  rewrites: {
    'README.md': 'index.md',
  },
  srcExclude: ['api', 'cmd', 'examples', 'docker', 'config', 'meta', 'out', 'util', 'scripts', 'build', 'buildkit', '**/TODO.md'],
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
    ],
    logo: 'https://gitlab.wikimedia.org/repos/releng/blubber/-/raw/main/docs/logo-400.png',
    docFooter: {
      prev: false,
      next: false
    },
    sidebar: [
      {
        text: 'Documentation',
        items: [
          { text: 'Configuration',
            link: '/configuration',
            items: [
              {text: 'Variants', link: '/configuration#variants'},
              {text: 'APT', link: '/configuration#apt'},
              {text: 'NodeJS', link: '/configuration#node'},
              {text: 'PHP', link: '/configuration#php'},
              {text: 'Python', link: '/configuration#python'},
            ]}
        ],
      },
      {
        text: 'Development',
        items: [
          { text: 'Changelog', link: '/CHANGELOG'},
          { text: 'Code', link: 'https://gitlab.wikimedia.org/repos/releng/blubber'},
          { text: 'Contributing', link: '/CONTRIBUTING'},
          { text: 'Release', link: '/RELEASE'}
        ]
      },
      {
        text: 'Other',
        collapsed: 'true',
        items: [
          { text: 'Deploying services to production', link: 'https://www.mediawiki.org/wiki/GitLab/Workflows/Deploying_services_to_production'},
        ]
      }
    ],

    socialLinks: [
    ],

    search: {
      provider: 'local'
    }
  }
})
