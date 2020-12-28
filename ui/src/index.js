import React from 'react'
import ReactDOM from 'react-dom'
import 'bootstrap/dist/css/bootstrap.min.css'
import './theme.css'
import reportWebVitals from "./reportWebVitals"
import Footer from './footer'
import Navbar from './navbar'
import About from './about'
import Parts from './parts'
import {Banner, YoutubeIframe} from './utils'

const axios = require('axios').default;

class Index extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            project: {}
        }
    }

    componentDidMount() {
        axios.get('/projects_api', {params: {latest: true}})
            .then(response => this.setState({project: response.data[0]}))
    }

    render() {
        return <div className="container">
            <div className="row row-cols-1 justify-content-md-center text-center m-2">
                <div className="col">
                    <Banner YoutubeLink={this.state.project.YoutubeLink} BannerLink={this.state.project.BannerLink}/>
                </div>
                <div className="col">
                    <YoutubeIframe YoutubeEmbed={this.state.project.YoutubeEmbed}/>
                </div>
            </div>
            <div className="row justify-content-md-center text-center m-2">
                <div className="col text-center mt-2">
                    <p>
                        If you would like to join our orchestra or get more information about our current projects,
                        please join us on <a href="https://discord.gg/9RVUJMQ">Discord!</a>
                    </p>
                </div>
            </div>
        </div>
    }
}


class Main extends React.Component {
    render() {
        switch (window.location.pathname) {
            case "/":
                return <Index/>
            case "/about":
                return <About/>
            case "/parts":
                return <Parts/>
            default:
                return 404
        }
    }
}

ReactDOM.render(
    <div>
        <Navbar/>
        <Main/>
        <Footer/>
    </div>,
    document.getElementById('root')
)

// ref: https://bit.ly/CRA-vitals
reportWebVitals(console.log);
