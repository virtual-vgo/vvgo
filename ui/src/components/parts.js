import React from 'react'
import MaterialTable from "material-table";
import {ProjectBanner} from "./utils";
import {Link as RouteLink, Route, Switch, useRouteMatch} from "react-router-dom";
import Button from "@material-ui/core/Button";
import tableIcons from "./table_icons";
import Container from "@material-ui/core/Container";
import ButtonGroup from "@material-ui/core/ButtonGroup";

export default function Parts(props) {
    let {path, url} = useRouteMatch();

    function PartsNav() {
        return props.projects.map(project =>
            <Button color='primary' component={RouteLink} key={project.Name} to={`${url}/${project.Name}`}>
                {project.Title}
            </Button>)
    }

    return <Container>
        <Switch>
            <Route exact path={path}>
                <PartsNav/>
                <h3>Please select a topic.</h3>
            </Route>
            {props.projects.map(project => <Route key={project.Name} path={`${path}/${project.Name}`}>
                <PartsNav/>
                <PartsTab project={project} parts={props.parts}/>
            </Route>)}
        </Switch>
    </Container>
}

function PartsTab(props) {
    let wantParts = []
    props.parts.forEach(part => {
        if (props.project.Name === part.Project) {
            wantParts.push(part)
        }
    })

    return <div>
        <div className="row">
            <div className="col mt-3 text-center">
                <WarnIf condition={props.project.PartsArchived === true}>
                    This project has been archived. Parts are only visible to leaders.
                </WarnIf>
                <WarnIf condition={props.project.PartsReleased === false}>
                    This project is unreleased and invisible to members!
                </WarnIf>
                <ProjectBanner project={props.project}/>
                <ProjectInfo project={props.project}/>
            </div>
        </div>
        <div className="row justify-content-center">
            <div className="col-auto">
                <ProjectLinks {...props.project}/>
            </div>
        </div>
        <div className="row justify-content-center">
            <div className="col mt-4">
                <PartsTable parts={wantParts}/>
            </div>
        </div>
    </div>
}

function WarnIf(props) {
    if (props.condition) {
        return <Warning>{props.children}</Warning>
    } else {
        return null
    }
}

function Warning(props) {
    return <div className="alert alert-warning">{props.children}</div>
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
            <h4><strong>Submission Deadline: </strong>
                <em>{props.project.SubmissionDeadline} (Hawaii Time)</em></h4>
        </div>
    </div>
}

function ProjectLinks(props) {
    return <ButtonGroup variant="text" color="primary">
        <Button href="https://www.youtube.com/watch?v=VgqtZ30bMgM">
            <i className="fab fa-youtube"/> Recording Instructions
        </Button>
        <Button href={props.ReferenceTrack}>
            <i className="far fa-file-audio"/> Reference Track
        </Button>
        <Button href={props.SubmissionLink}>
            <i className="fab fa-dropbox"/> Submit Recordings
        </Button>
    </ButtonGroup>
}

function PartsTable(props) {
    function PartTitle(props) {
        return <div className="title text-left text-nowrap">
            {props.children}
        </div>
    }

    return <MaterialTable
        tableIcons={tableIcons}
        columns={[
            {
                title: <h4>Parts</h4>,
                field: "PartName",
                render: rowData => <PartTitle>{rowData.PartName}</PartTitle>
            },
            {
                title: <h4>Downloads</h4>,
                render: rowData => <PartDownloads part={rowData}/>,
                searchable: false,
                sorting: false
            }
        ]}
        data={props.parts}
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



