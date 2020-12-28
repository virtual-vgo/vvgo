import React from 'react'
import ReactDOM from 'react-dom'
import 'bootstrap/dist/css/bootstrap.min.css'
import './theme.css'
import reportWebVitals from "./reportWebVitals"
import Footer from './footer'
import Navbar from './navbar'

const axios = require('axios').default;

function Banner(props) {
    return <a href={props.YoutubeLink} className="btn btn-link nav-link">
        <img src={props.BannerLink}
             className="mx-auto img-fluid"
             alt="banner"/>
    </a>;
}

function YoutubeIframe(props) {
    return <div className="project-iframe-wrapper text-center m-2">
        <iframe className="project-iframe" src={props.YoutubeEmbed}
                allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
                title="latest.Title"
                allowFullScreen/>
    </div>;
}

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
        return <div className="mt-2 container">
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

ReactDOM.render(
    <div>
        <Navbar/>
        <Index/>
        <Footer/>
    </div>,
    document.getElementById('root'));

// ref: https://bit.ly/CRA-vitals
reportWebVitals(console.log);
