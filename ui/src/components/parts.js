import React from 'react'
import MaterialTable from "material-table";
import {ProjectBanner} from "./utils";

import {Link, Route, Switch, useRouteMatch} from "react-router-dom";
import Button from "@material-ui/core/Button";

const axios = require('axios').default;

class Parts extends React.Component {
    constructor(props) {
        super(props)
        this.state = {projects: []}
    }

    componentDidMount() {
        axios.get('/projects_api').then(response => this.setState({projects: response.data}))
    }

    render() {
        return <div className="container">
            <ProjectsNavbar projects={this.state.projects}/>
        </div>
    }
}

function ProjectsNavbar(props) {
    let {path, url} = useRouteMatch();
    return <div>
        {props.projects.map(project => <Button key={project.Name}>
            <Link className="nav-link" to={`${url}/${project.Name}`}>{project.Title}</Link>
        </Button>)}
        <Switch>
            <Route exact path={path}>
                <h3>Please select a topic.</h3>
            </Route>
            {props.projects.map(project => <Route path={`${path}/${project.Name}`}>
                <PartsTab project={project} key={project.Name}/>
            </Route>)}
        </Switch>
    </div>
}

export default Parts

function PartsTab(props) {
    return <div>
        <div className="row">
            <div className="col mt-3 text-center">
                <ArchivedWarning project={props.project}/>
                <UnreleasedWarning project={props.project}/>
                <ProjectBanner project={props.project}/>
                <ProjectInfo project={props.project}/>
            </div>
        </div>
        <div className="row justify-content-center">
            <div className="col-auto">
                <ProjectLinks project={props.project}/>
            </div>
        </div>
        <div className="row justify-content-center">
            <div className="col mt-4">
                <PartsTable Project={props.project.Name}/>
            </div>
        </div>
    </div>
}

function ArchivedWarning(props) {
    if (props.project.PartsArchived) {
        return <div className="alert alert-warning">
            This project has been archived. Parts are only visible to leaders.
        </div>
    } else {
        return null
    }
}

function UnreleasedWarning(props) {
    if (!props.project.PartsReleased) {
        return <div className="alert alert-warning">
            This project is unreleased and invisible to members!
        </div>
    } else {
        return null
    }
}

function ProjectInfo(props) {
    return <div className="row row-cols-1">
        <div className="col text-center">
            {props.project.Composers}<br/>
            <small>{props.project.Arrangers}</small>
        </div>
        <div className="col text-center">
            <a href={props.project.PartsLink} className="text-light">link to parts <i className="fas fa-link"/></a>
        </div>
        <div className="col text-center m-2">
            <h4><strong>Submission Deadline:</strong>
                <em>{props.project.SubmissionDeadline} (Hawaii Time)</em></h4>
        </div>
    </div>
}

function ProjectLinks(props) {
    let cardClass = "card bg-transparent text-center"
    let cardRefClass = "btn btn-lnk btn-outline-light text-info"
    return <div className="card-deck">
        <div className={cardClass}>
            <a className={cardRefClass} href="https://www.youtube.com/watch?v=VgqtZ30bMgM">
                <i className="fab fa-youtube"/> Recording Instructions
            </a>
        </div>
        <div className={cardClass}>
            <a className={cardRefClass} href={props.project.ReferenceTrack}>
                <i className="far fa-file-audio"/> Reference Track
            </a>
        </div>
        <div className={cardClass}>
            <a className={cardRefClass} href={props.project.SubmissionLink}>
                <i className="fab fa-dropbox"/> Submit Recordings
            </a>
        </div>
    </div>
}


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

    render() {
        return <MaterialTable
            columns={[
                {
                    title: <h4>Parts</h4>,
                    field: "PartName",
                    render: rowData => <PartTitle PartName={rowData.PartName}/>
                },
                {
                    title: <h4>Downloads</h4>,
                    render: rowData => <PartDownloads part={rowData}/>,
                    searchable: false,
                    sorting: false
                }
            ]}
            data={this.state.parts}
            options={{
                showTitle: false, paging: false, isLoading: true, padding: "dense",
                searchFieldStyle: {
                    color: "black",
                    backgroundColor: "white"
                },
                headerStyle: {
                    color: "white",
                    backgroundColor: "inherit",
                }
            }}
            style={{
                width: "100%",
                color: "white",
                backgroundColor: "inherit",
                maxWidth: "800px",
                margin: "auto"
            }}
        />
    }
}

function PartTitle(props) {
    return <div className="title text-left text-nowrap">{props.PartName}</div>
}

function PartDownloads(props) {
    return <div className="row justify-content-start">
        <ContentLink url={props.part.SheetMusicLink}>
            <i className="far fa-file-pdf"/> sheet music
        </ContentLink>
        <ContentLink url={props.part.ClickTrackLink}>
            <i className="far fa-file-audio"/> click track
        </ContentLink>
        <ContentLink url={props.part.ConductorVideo}>
            <i className="far fa-file-video"/> conductor video
        </ContentLink>
        <ContentLink url={props.part.PronunciationGuideLink}>
            <i className="fas fa-language"/> pronunciation guide
        </ContentLink>
    </div>
}

function ContentLink(props) {
    if (props.url !== "") {
        return <div className="col text-nowrap">
            <Button><a className="text-light" href={props.url}>{props.children}</a></Button>
        </div>
    } else {
        return null
    }
}



