const { override, addWebpackAlias, addWebpackPlugin } = require('customize-cra');
const webpack = require('webpack');

module.exports = override(
  addWebpackAlias({
    'http': 'stream-http',
    'https': 'https-browserify',
    'util': 'util',
    'zlib': 'browserify-zlib',
    'process': 'process/browser',
    'stream': 'stream-browserify',
    'buffer': 'buffer',
    'asset': 'assert'
  }),
  addWebpackPlugin(
    new webpack.ProvidePlugin({
      process: 'process/browser',
      Buffer: ['buffer', 'Buffer']
    })
  ),
  (config) => {
    if (!config.resolve) {
      config.resolve = {};
    }
    if (!config.resolve.alias) {
      config.resolve.alias = {};
    }
    Object.assign(config.resolve.alias, {
      process: "process/browser",
      zlib: "browserify-zlib",
      stream: "stream-browserify",
      util: "util",
      buffer: "buffer",
      asset: "assert"
    });
    return config;
  }
);
