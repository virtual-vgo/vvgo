module.exports = {
    compact: true,
    presets: [
        ["@babel/env", {targets: {node: 6}}],
        ["@babel/preset-react", {runtime: "automatic"}],
        ["@babel/preset-typescript"]
    ],
    plugins: ["lodash"]
};
