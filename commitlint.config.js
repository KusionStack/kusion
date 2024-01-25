// This file is the configuration file of [commitlint](https://commitlint.js.org/#/).
//
// Rules can be referenced: 
// https://github.com/conventional-changelog/commitlint/tree/master/%40commitlint/config-conventional
const Configuration = {
    extends: ['@commitlint/config-conventional'],

    rules: {
        'body-max-line-length': [2, 'always', 500],
    },
};

module.exports = Configuration;
