const path = require('path')
const webpack = require('webpack')

module.exports = {
    entry: {
        feature: './js/src/feature.js',
        index: './js/src/index.js'
    },
    mode: 'development',
    devtool: false,
    module: {
        rules: [
            {
                test: /\.js$/,
                exclude: /node_modules/,
                loader: 'babel-loader',
                options: {presets: ['@babel/preset-react']}
            },
            {test: /\.css$/, use: ['style-loader', 'css-loader']}
        ]
    },
    resolve: {extensions: ['*', '.js']},
    output: {
        path: path.resolve(__dirname, 'dist/'),
        library: 'Bundle'
    }
}
