import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';

export const gitConfig = {
  user: 'dr3dnought',
  repo: 'dew',
  branch: 'main',
};

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: 'Dew',
    },
    links: [
      { text: 'Get Started', url: '/docs' },
      { text: 'Queries', url: '/docs/select' },
      { text: 'Advanced', url: '/docs/transactions' },
      { text: 'Examples', url: '/docs/examples-patterns' },
    ],
    githubUrl: `https://github.com/${gitConfig.user}/${gitConfig.repo}`,
  };
}
