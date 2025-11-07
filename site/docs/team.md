---
layout: page
---
<script setup>
import {
  VPTeamPage,
  VPTeamPageTitle,
  VPTeamMembers
} from 'vitepress/theme';

const members = [
  {
    avatar: 'https://www.github.com/zergslaw.png',
    name: 'Edgar Sipki',
    title: 'Founder',
    links: [
      { icon: 'github', link: 'https://github.com/zergslaw' },
    ]
  },
  {
    avatar: 'https://www.github.com/hound672.png',
    name: 'Vasilii Bliznetsov',
    title: 'Founder',
    links: [
      { icon: 'github', link: 'https://github.com/hound672' },
    ]
  },
  {
    avatar: 'https://www.github.com/onokonem.png',
    name: 'Daniel Podolsky',
    title: 'Founder',
    links: [
      { icon: 'github', link: 'https://github.com/onokonem' },
    ]
  },
  {
    avatar: 'https://github.com/Yakwilik.png',
    name: 'Khasbulat Abdullin',
    title: 'Contributor',
    links: [
      { icon: 'github', link: 'https://github.com/Yakwilik' },
    ]
  }
]
</script>

<VPTeamPage>
  <VPTeamPageTitle>
    <template #title>
      Our Team
    </template>
    <template #lead>
      The founding and development of EasyP are driven by an international team, all of whom are featured below.
    </template>
  </VPTeamPageTitle>
  <VPTeamMembers
    :members="members"
  />
</VPTeamPage>
