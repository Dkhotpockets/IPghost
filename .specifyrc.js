const publicPath = './output';

/*
 * Finally exports the configuration
 */
module.exports = {
  // Find more about how to target a Specify repository at: https://docs.specifyapp.com/reference/parsers-engine
  repository: '@owner/name',
  personalAccessToken: '<your-personal-access-token>',
  rules: [
    {
      name: 'Extract all tokens to JSON',
      path: `${publicPath}/all-tokens.json`,
      parsers: [],
    },
  ],
};
