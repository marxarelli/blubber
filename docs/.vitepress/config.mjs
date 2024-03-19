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
    logo: '/logo-400.png',
    docFooter: {
      prev: false,
      next: false
    },
    sidebar: [
      {
        text: 'Documentation',
        items: [
          { text: 'Home',
            link: '/',
            items: [
              {text: 'Examples', link: '/#examples'},
              {text: 'Concepts', link: '/#concepts'},
              {text: 'Usage', link: '/#usage'}]
          },
          { text: 'Configuration',
            link: '/configuration',
            items: [
              {text: 'Variants', link: '/configuration#variants'},
              {text: 'APT', link: '/configuration#apt'},
              {text: 'NodeJS', link: '/configuration#node-1'},
              {text: 'PHP', link: '/configuration#php-1'},
              {text: 'Python', link: '/configuration#python-1'},
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
      }
    ],

    search: {
      provider: 'local'
    }
  }
})
