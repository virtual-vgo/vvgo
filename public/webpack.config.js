const path = require('path')
const ESLintPlugin = require('eslint-webpack-plugin');

module.exports = {
    entry: {
        index: './js/src/index.tsx',
    },
    mode: 'development',
    devtool: 'inline-source-map',
    module: {
        rules: [
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
            },
            {test: /\.css$/, use: ['style-loader', 'css-loader']}
        ]
    },
    plugins: [new ESLintPlugin()],
    resolve: {extensions: ['*', '.js', '.ts', '.tsx']},
    output: {
        path: path.resolve(__dirname, 'dist/'),
        filename: 'index.js'
    }
}
