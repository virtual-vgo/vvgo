const path = require('path')
const ESLintPlugin = require('eslint-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const devApi = 'http://localhost:42069'

module.exports = {
    entry: {index: './ui/src/index.js'},
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
                test: /\.tsx?$/,
                exclude: /node_modules/,
                use: 'ts-loader',
            }
        ]
    },
    optimization: {splitChunks: {}},
    plugins: [
        new ESLintPlugin(),
        new HtmlWebpackPlugin({
            title: 'VVGO | Virtual Video Game Orchestra',
            template: "./ui/src/index.html"
        })],
    resolve: {extensions: ['.ts', '.tsx', '...']},
    output: {
        path: path.resolve(__dirname, './public/dist'),
        filename: './index.js',
        clean: true,
        publicPath: "/"
    },
    devServer: {
        hot: true,
        liveReload: false,
        static: false,
        proxy: {'/api': devApi, '/images': devApi, '/download': devApi},
        host: 'localhost',
        port: 8080,
        historyApiFallback: true,
    },
}
