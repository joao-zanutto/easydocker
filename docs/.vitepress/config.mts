import {defineConfig} from "vitepress";
import {readFileSync} from "node:fs";
import {dirname, resolve} from "node:path";
import {fileURLToPath} from "node:url";

const base = "/easydocker/";
const __dirname = dirname(fileURLToPath(import.meta.url));

function slugify(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^\w\- ]+/g, "")
    .replace(/\s+/g, "-")
    .replace(/\-+/g, "-");
}

function getFeatureHeadings(): Array<{text: string; link: string}> {
  const featuresPath = resolve(__dirname, "../features.md");
  const raw = readFileSync(featuresPath, "utf8");
  const entries: Array<{text: string; link: string}> = [];
  const headingRe = /^##\s+(.*)$/gm;
  let match;

  while ((match = headingRe.exec(raw)) !== null) {
    const title = match[1].trim();
    if (!title) continue;
    entries.push({
      text: title,
      link: `/features#${slugify(title)}`,
    });
  }

  return entries;
}

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
      {text: "Features", link: "/features"},
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
            {
              text: "Features",
              link: "/features",
              items: getFeatureHeadings(),
            },
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
