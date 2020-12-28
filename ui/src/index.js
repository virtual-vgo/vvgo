import React from 'react';
import ReactDOM from 'react-dom';
import 'bootstrap/dist/css/bootstrap.min.css';
import './theme.css';
import reportWebVitals from "./reportWebVitals";

const axios = require('axios').default;
const instance = axios.create({
    baseURL: 'http://localhost:8080',
    timeout: 1000,
});

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
            .catch(function (error) {
                if (error.response) {
                    // The request was made and the server responded with a status code
                    // that falls out of the range of 2xx
                    console.log(error.response.data);
                    console.log(error.response.status);
                    console.log(error.response.headers);
                } else if (error.request) {
                    // The request was made but no response was received
                    // `error.request` is an instance of XMLHttpRequest in the browser and an instance of
                    // http.ClientRequest in node.js
                    console.log(error.request);
                } else {
                    // Something happened in setting up the request that triggered an Error
                    console.log('Error', error.message);
                }
                console.log(error.config);
            })
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
    <Index/>,
    document.getElementById('root')
);

// ref: https://bit.ly/CRA-vitals
reportWebVitals(console.log);
