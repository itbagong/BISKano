module.exports = {
  presets: ['module:metro-react-native-babel-preset'],
  plugins: [
    [
      'module-resolver',
      {
        root: ['./src'],
        extensions: [
          '.ios.js',
          '.android.js',
          '.js',
          '.ts',
          '.tsx',
          '.json',
          '.png',
        ],
        alias: {
          '@components': './src/components',
          '@navigations': './src/navigations',
          '@screens': './src/screens',
          '@images': './src/assets/images',
          '@assets': './src/assets',
          '@utils': './src/utils',
          '@data': './src/data',
          '@overmind': './src/overmind',
          overmind: 'overmind',
        },
      },
    ],
    [
      'module:react-native-dotenv',
      {
        envName: 'APP_ENV',
        moduleName: '@env',
        path: '.env',
        blocklist: null,
        allowlist: null,
        safe: false,
        allowUndefined: true,
        verbose: false,
      },
    ],
  ],
};
