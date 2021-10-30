const path = require('path')
const ESLintPlugin = require('eslint-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const vvgoApi = 'http://localhost:8080'

module.exports = {
    entry: {index: './src/index.js'},
    mode: 'development',
    devtool: 'inline-source-map',
    module: {
        rules: [
            {
                test: /\.css$/i,
                use: ["style-loader", "css-loader"],
            }, {
                test: /\.s[ac]ss$/i,
                use: ["style-loader", "css-loader", "sass-loader",],
            }, {
                test: /\.(png|svg|jpg|jpeg|gif)$/i,
                type: 'asset/resource',
            }, {
                test: /\.js$/,
                exclude: /node_modules/,
                loader: 'babel-loader',
                options: {presets: ['@babel/preset-react']}
            }, {
                test: /\.tsx?$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            }
        ]
    },
    optimization: {splitChunks: {}},
    plugins: [
        new ESLintPlugin(),
        new HtmlWebpackPlugin({
            title: 'VVGO | Virtual Video Game Orchestra',
            template: "./src/index.html"
        })],
    resolve: {extensions: ['.ts', '.tsx', '...']},
    output: {
        path: path.resolve(__dirname, './dist'),
        filename: './index.js',
        clean: true,
        publicPath: "/"
    },
    devServer: {
        hot: true,
        liveReload: false,
        static: false,
        proxy: {'/api': vvgoApi, '/images': vvgoApi},
        port: 8081,
        historyApiFallback: true,
    },
}
