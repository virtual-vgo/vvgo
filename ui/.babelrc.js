module.exports = {
    compact: true,
    presets: [
        ["@babel/env", {"targets": {"node": 6}}],
    ],
    plugins: [
        "lodash",
        ['babel-plugin-direct-import', {modules: ['@mui/material', '@mui/icons-material']}],
    ]
};
