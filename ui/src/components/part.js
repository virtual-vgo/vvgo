import React from "react";
import {ButtonGroup, Typography} from "@material-ui/core";
import Button from "@material-ui/core/Button";
import Box from "@material-ui/core/Box";
import {YoutubeIframe} from "./utils";
import Container from "@material-ui/core/Container";

export default function Part(props) {
    function SheetMusic() {
        if (props.part.SheetMusicLink !== "") {
            return <Box>
                <object style={{width: '100%', height: '90vh'}} data={props.part.SheetMusicLink}/>
            </Box>
        } else {
            return null
        }
    }

    function ClickTrack() {
        if (props.part.ClickTrackLink !== "") {
            return <object data={props.part.ClickTrackLink}/>
        } else {
            return null
        }
    }

    function ConductorVideo() {
        if (props.part.ConductorVideo !== "") {
            return <Box width={'50%'}>
                <YoutubeIframe YoutubeEmbed="https://www.youtube.com/embed/bsFMaH1tTws"/>
            </Box>
        } else {
            return null
        }
    }

    props.setAppTitle(`${props.project.Title} | ${props.part.PartName}`)
    document.title = `${props.project.Title} | ${props.part.PartName}`
    console.log("displaying", props.part)
    return <Container>
        <ButtonGroup variant='outlined'>
            <ProjectLinks {...props.project}/>
            <PartDownloads {...props.part}/>
        </ButtonGroup>
        <ConductorVideo/>
        <SheetMusic/>
    </Container>
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
