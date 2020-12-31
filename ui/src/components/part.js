import React from "react";
import {ButtonGroup, Typography} from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Container from "@material-ui/core/Container";
import VVGOAppBar from "./app_bar";

export default function Part(props) {
    function SheetMusic() {
        if (props.part.SheetMusicLink !== "") {
            return <div>
                Sheet Music: <embed style={{width: '100%', height: '90vh'}} src={props.part.SheetMusicLink}/>
            </div>
        } else {
            return null
        }
    }

    function ClickTrack() {
        if (props.part.ClickTrackLink !== "") {
            return <div>
                Click Track: <audio controls src={props.part.ClickTrackLink}/>
            </div>
        } else {
            return null
        }
    }

    function ReferenceTrack() {
        if (props.project.ReferenceTrack !== "") {
            return <div>
                Reference Track: <audio controls src={props.project.ReferenceTrack}/>
            </div>
        } else {
            return null
        }
    }

    document.title = `${props.project.Title} | ${props.part.PartName}`
    console.log("displaying", props.part)
    return <div>
        <VVGOAppBar drawerState={props.drawerState} title={`${props.project.Title} | ${props.part.PartName}`}/>
        <Container>
            <ProjectInfo {...props.project}/>
            <ButtonGroup variant='outlined'>
                <ProjectLinks {...props.project}/>
                <PartDownloads {...props.part}/>
            </ButtonGroup>
            <ReferenceTrack/>
            <ClickTrack/>
            <SheetMusic/>
        </Container>
    </div>
}

function ProjectInfo(props) {
    return <div>
        <Typography variant='h4'>
            <strong>Submission Deadline:</strong> <em>{props.SubmissionDeadline} (Hawaii Time)</em>
        </Typography>
    </div>
}

function ProjectLinks(props) {
    return [
        {
            href: "https://www.youtube.com/watch?v=VgqtZ30bMgM",
            startIcon: <i className="fab fa-youtube"/>,
            children: 'Recording Instructions'
        },
        {
            href: props.SubmissionLink,
            startIcon: <i className="fab fa-dropbox"/>,
            children: 'Submit Recordings'
        },
        {
            href: props.ReferenceTrack,
            startIcon: <i className="far fa-file-audio"/>,
            children: 'Reference Track'
        },
    ].map(button => <Button key={button.href} {...button}/>)
}

function PartDownloads(props) {
    return [
        {
            href: props.SheetMusicLink,
            startIcon: <i className="far fa-file-pdf"/>,
            children: 'sheet music'
        },
        {
            href: props.ClickTrackLink,
            startIcon: <i className="far fa-file-audio"/>,
            children: 'click track'
        },
        {
            href: props.ConductorVideo,
            startIcon: <i className="far fa-file-video"/>,
            children: 'conductor video'
        },
        {
            href: props.PronunciationGuideLink,
            startIcon: <i className="fas fa-language"/>,
            children: 'pronunciation guide'
        },
    ].filter(b => b.href !== "").map(button => <Button key={button.href} {...button}/>)
}
