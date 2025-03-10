const { override, addWebpackPlugin } = require('customize-cra');
const webpack = require('webpack');

module.exports = override(
  (config) => {
    // Configure aliases for polyfills
    config.resolve.alias = {
      ...config.resolve.alias,
      'stream': require.resolve('stream-browserify'),
      'crypto': require.resolve('crypto-browserify'),
      'http': require.resolve('stream-http'),
      'https': require.resolve('https-browserify'),
      'os': require.resolve('os-browserify/browser'),
      'process': require.resolve('browser-process-hrtime'),
    };

    // Add necessary plugins
    config.plugins = (config.plugins || []).concat([
      new webpack.ProvidePlugin({
        process: 'browser-process-hrtime',
        Buffer: ['buffer', 'Buffer']
      })
    ]);

    return config;
  },
  addWebpackPlugin(
    new webpack.ProvidePlugin({
      process: 'browser-process-hrtime',
    }),
  )
);
