module.exports = {
    compact: true,
    presets: [
        ["@babel/env", {targets: {esmodules: true}}],
        ["@babel/preset-react", {runtime: "automatic"}],
        ["@babel/preset-typescript"]
    ],
    plugins: ["lodash"]
};
