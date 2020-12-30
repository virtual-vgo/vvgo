import React from "react";
import Button from "@material-ui/core/Button";
import Helmet from "react-helmet";
import {ButtonGroup, Typography} from "@material-ui/core";
import ReactPlayer from "react-player";

export default function Part(props) {
    return <div>
        <Helmet>
            <title>{props.project.Title} | {props.part.PartName}</title>
            <meta name="description" content=""/>
        </Helmet>
        <Typography variant='h3'>{props.project.Title} - {props.part.PartName}</Typography>
        <ProjectInfo {...props.project}/>
        <ButtonGroup variant='outlined'>
            <ProjectLinks {...props.project}/>
            <PartDownloads {...props.part}/>
        </ButtonGroup>
        <ReactPlayer url={props.part.ReferenceTrack}/>
    </div>
}

function ProjectInfo(props) {
    return <div>
        <Typography paragraph>
            {props.Composers}<br/>
            <small>{props.Arrangers}</small>
        </Typography>
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
            href: props.ReferenceTrack,
            startIcon: <i className="far fa-file-audio"/>,
            children: 'Reference Track'
        },
        {
            href: props.SubmissionLink,
            startIcon: <i className="fab fa-dropbox"/>,
            children: 'Submit Recordings'
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
