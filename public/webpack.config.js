const path = require('path')
const ESLintPlugin = require('eslint-webpack-plugin');

module.exports = {
    entry: {index: './js/src/index.js'},
    mode: 'development',
    devtool: 'inline-source-map',
    module: {
        rules: [
            {
                test: /\.css$/i,
                use: ["style-loader", "css-loader"],
            },
            {
                test: /\.s[ac]ss$/i,
                use: ["style-loader", "css-loader", "sass-loader",],
            },
            {
                test: /\.js$/,
                exclude: /node_modules/,
                loader: 'babel-loader',
                options: {presets: ['@babel/preset-react']}
            },
            {
                test: /\.tsx?$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            }
        ]
    },
    plugins: [new ESLintPlugin()],
    resolve: {extensions: ['*', '.js', '.ts', '.tsx']},
    output: {
        path: path.resolve(__dirname, 'dist/'),
        filename: 'index.js'
    }
}
