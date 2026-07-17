import { Config } from '@remotion/cli/config';

Config.setEntryPoint('src/index.ts');
Config.setVideoImageFormat('jpeg');
Config.setOverwriteOutput(true);

// The monorepo hoists frontend's TypeScript 7 (native preview, no `ts.sys`),
// which crashes Remotion's esbuild-loader when it reads tsconfig.json.
// Providing tsconfigRaw makes the loader skip require('typescript') entirely.
const tsconfigRaw = {
  compilerOptions: {
    jsx: 'react-jsx',
    target: 'ES2022',
  },
};

Config.overrideWebpackConfig((config) => {
  for (const rule of config.module?.rules ?? []) {
    if (!rule || typeof rule !== 'object' || !('use' in rule)) continue;
    const uses = Array.isArray(rule.use) ? rule.use : [rule.use];
    for (const use of uses) {
      if (
        use &&
        typeof use === 'object' &&
        'loader' in use &&
        typeof use.loader === 'string' &&
        use.loader.includes('esbuild') &&
        use.options &&
        typeof use.options === 'object'
      ) {
        (use.options as Record<string, unknown>).tsconfigRaw = tsconfigRaw;
      }
    }
  }
  return config;
});
