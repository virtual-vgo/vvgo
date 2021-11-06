module.exports = {
    env: {browser: true, es6: true},
    extends: [
        "eslint:recommended",
        "plugin:react/recommended",
        "plugin:react/jsx-runtime",
        "plugin:@typescript-eslint/recommended"
    ],
    parser: "@typescript-eslint/parser",
    parserOptions: {
        sourceType: "module",
        ecmaFeatures: {jsx: true, impliedStrict: true}
    },
    plugins: ["react", "@typescript-eslint"],
    settings: {react: {version: "detect"}},
    rules: {
        "react/jsx-uses-react": 1,
        "no-unneeded-ternary": 1,
        "no-var": 1,
    }
};
