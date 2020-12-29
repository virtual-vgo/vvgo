import React from 'react'
import {Banner} from "./utils";

const axios = require('axios').default;

class PartsTable extends React.Component {
    constructor(props) {
        super(props)
        this.state = {parts: [], done: false}
    }

    fetchParts() {
        return axios.get('/parts_api', {params: {project: this.props.Project}})
    }

    componentDidMount() {
        this.fetchParts()
            .then(response => {
                this.setState({parts: response.data, done: true})
            })
            .catch(error => console.log(error))
    }

    contentLink(url, info) {
        const colClass = "col-sm-auto text-nowrap"
        const downloadClass = "btn btn-sm btn-link btn-outline-light bg-dark text-light"
        if (url !== "") {
            return <div className={colClass}>
                <a href={url} className={downloadClass}>{info}</a>
            </div>
        }
    }

    partRow(part) {
        let sheetMusic = this.contentLink(part.SheetMusicLink,
            <div><i className="far fa-file-pdf"/> sheet music</div>)

        let clickTrack = this.contentLink(part.ClickTrackLink,
            <div><i className="far fa-file-audio"/> click track</div>)

        let conductorVideo = this.contentLink(part.ClickTrackLink,
            <div><i className="far fa-file-video"/> conductor video</div>)

        let pronunciationGuide = this.contentLink(part.PronunciationGuideLink,
            <div><i className="fas fa-language"/> pronunciation guide</div>)

        return <tr key={part.PartName}>
            <td>{part.ScoreOrder}</td>
            <td className="title text-nowrap">{part.PartName}</td>
            <td>
                <div className="row justify-content-start">
                    {sheetMusic}
                    {clickTrack}
                    {conductorVideo}
                    {pronunciationGuide}
                </div>
            </td>
        </tr>
    }

    render() {
        if (this.state.done) {
            let rows = this.state.parts.map(part => this.partRow(part))
            return <table className="table dt-responsive text-light w-100">
                <thead>
                <tr>
                    <th>Score Order</th>
                    <th>Part</th>
                    <th>Downloads</th>
                </tr>
                </thead>
                <tbody>{rows}</tbody>
            </table>
        } else {
            return <div>parts loading</div>
        }
    }
}

class Parts extends React.Component {
    projectHeader(project) {
        let archivedWarning = null
        if (project.PartsArchived) {
            archivedWarning = <div className="alert alert-warning">
                This project has been archived. Parts are only visible to leaders.
            </div>
        }

        let unreleasedWarning = null
        if (!project.PartsReleased) {
            unreleasedWarning = <div className="alert alert-warning">
                This project is unreleased and invisible to members!
            </div>
        }

        let banner = <Banner YoutubeLink={project.YoutubeLink} BannerLink={project.BannerLink}/>
        if (project.YoutubeLink === "") {
            banner = <div>
                <h2 className="title">{project.Title}</h2>
                <h3>{project.Sources}</h3>
            </div>
        }
        return <div>
            {archivedWarning}
            {unreleasedWarning}
            {banner}
            <div className="row row-cols-1">
                <div className="col text-center">
                    {project.Composers}<br/>
                    <small>{project.Arrangers}</small>
                </div>
                <div className="col text-center">
                    <a href={project.PartsLink} className="text-light">link to parts <i className="fas fa-link"/></a>
                </div>
                <div className="col text-center m-2">
                    <h4><strong>Submission Deadline:</strong>
                        <em>{project.SubmissionDeadline} (Hawaii Time)</em></h4>
                </div>
            </div>
        </div>
    }

    projectLinks(project) {
        let cardClass = "card bg-transparent text-center"
        let cardRefClass = "btn btn-lnk btn-outline-light text-info"
        return <div className="card-deck">
            <div className={cardClass}>
                <a className={cardRefClass} href="https://www.youtube.com/watch?v=VgqtZ30bMgM">
                    <i className="fab fa-youtube"/> Recording Instructions
                </a>
            </div>
            <div className={cardClass}>
                <a className={cardRefClass} href={project.ReferenceTrack}>
                    <i className="far fa-file-audio"/> Reference Track
                </a>
            </div>
            <div className={cardClass}>
                <a className={cardRefClass} href={project.SubmissionLink}>
                    <i className="fab fa-dropbox"/> Submit Recordings
                </a>
            </div>
        </div>
    }

    constructor(props) {
        super(props)
        this.state = {projects: []}
    }

    componentDidMount() {
        axios.get('/projects_api').then(response => this.setState({projects: response.data}))
    }

    render() {
        return this.state.projects.map(project => <div key={project.Name}>
            {this.projectHeader(project)}
            {this.projectLinks(project)}
            <PartsTable Project={project.Name}/>
        </div>)
    }
}

export default Parts
