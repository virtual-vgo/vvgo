const path = require('path')

module.exports = {
    entry: {
        index: './js/src/index.tsx',
        about: './js/src/about.tsx',
        sessions: './js/src/sessions.tsx',
        mixtape: './js/src/mixtape.tsx'
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
    resolve: {extensions: ['*', '.js', '.ts', '.tsx']},
    output: {
        path: path.resolve(__dirname, 'dist/'),
        library: 'Bundle'
    }
}
