import {defineConfig} from "vitepress";

const base = "/easydocker/";

export default defineConfig({
  base,
  title: "EasyDocker",
  description:
    "A Bubble Tea TUI for exploring Docker containers, images, networks, volumes, logs, and metrics.",
  head: [
    [
      "link",
      {rel: "icon", type: "image/png", href: `${base}easydocker-logo.png`},
    ],
  ],
  themeConfig: {
    logo: "/easydocker-logo.png",
    nav: [
      {text: "Install", link: "/install"},
      {text: "Usage", link: "/usage"},
      {text: "Changelog", link: "/changelog"},
      {text: "GitHub", link: "https://github.com/joao-zanutto/easydocker"},
    ],
    sidebar: {
      "/": [
        {
          text: "Getting started",
          items: [
            {text: "Home", link: "/"},
            {text: "Install", link: "/install"},
            {text: "Usage", link: "/usage"},
            {text: "Changelog", link: "/changelog"},
          ],
        },
      ],
    },
    socialLinks: [
      {icon: "github", link: "https://github.com/joao-zanutto/easydocker"},
    ],
  },
});
