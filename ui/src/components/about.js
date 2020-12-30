import React from 'react'
import Helmet from "react-helmet";
import {Container, Typography} from "@material-ui/core";

export default function About(props) {
    return <Container>
        <Helmet>
            <title>About</title>
            <meta name="description" content="About VVGO"/>
        </Helmet>
        <InfoRow/>
        <div className="row">
            <div className="col text-center"><h2>VVGO Leadership</h2></div>
        </div>
        <div className="row justify-content-md-center">
            <div className="col col-md-auto mt-4 text-center">
                <table className="table table-bordered table-responsive text-light">
                    <tbody>
                    {props.leaders.map(leader => <LeaderRow key={leader.Name} {...leader}/>)}
                    </tbody>
                </table>
            </div>
        </div>
    </Container>
}

function InfoRow() {
    return <div className="row mt-4 text-justify">
        <div className="col">
            <Typography paragraph={true} align={"justify"}>
                Formed in March 2020, <strong>Virtual Video Game Orchestra</strong> (VVGO, "vee-vee-go") is an
                online
                volunteer-run music ensemble predicated on providing a musical performance outlet for musicians
                whose
                IRL rehearsals and performances were cancelled due to COVID-19. Led and organized by members
                from
                various video game ensembles, and with a community of hundreds of musicians from across the
                globe,
                VVGO is open to any who wish to participate regardless of instrument, skill level, or musical
                background.
            </Typography>
            <p className="blockquote">
                Our mission is to provide a fun and accessible virtual community of musicians from around the
                world
                through performing video game music.
            </p>
            <p className="blockquote">
                We are always accepting new members into our community. If you would like to join our orchestra
                or
                get more information about our current performace opportunities, please join us on
                <a href="https://discord.gg/9RVUJMQ" className="text-info">Discord</a>!
            </p>
        </div>
    </div>
}

function LeaderRow(props) {
    let nameData = <p className="text-light">
        {props.Name}<br/><small>{props.Epithet}</small>
    </p>

    if (props.Email !== "") {
        let href = "mailto: " + props.Email
        nameData = <a className="text-light" href={href}>
            {props.Name}<br/><small>{props.Epithet}</small>
        </a>
    }

    return <tr key={props.Name}>
        <td><img src={props.Icon} alt={props.Name} height="125"/></td>
        <td>
            {nameData}
        </td>
        <td>
            <p>{props.Blurb}</p>
            <p><i>{props.Affiliations}</i></p>
        </td>
    </tr>
}
