import React = require("react");
import {RootContainer} from "./components";
import {Director, useDirectors} from "./datasets";

export const About = () => {
    const directors = useDirectors();

    return <RootContainer>
        <Blurb/>
        <Directors directors={directors}/>
    </RootContainer>;
};

const Blurb = () => {
    return <div className="row mt-4 border-bottom border-light">
        <div className="col col-lg-3 col-12 text-justify">
            <h2 className="text-center">About Us</h2>
        </div>
        <div className="col col-lg-9 col-12 text-justify fs-6">
            <p className="">
                Formed in March 2020, <strong>Virtual Video Game Orchestra</strong> (VVGO, &quot;vee-vee-go&quot;) is an
                online
                volunteer-run music ensemble predicated on providing a musical performance outlet for musicians
                whose
                IRL rehearsals and performances were cancelled due to COVID-19. Led and organized by members from
                various video game ensembles, and with a community of hundreds of musicians from across the globe,
                VVGO is open to any who wish to participate regardless of instrument, skill level, or musical
                background.
            </p>
            <p className="">
                Our mission is to provide a fun and accessible virtual community of musicians from around the world
                through performing video game music.
            </p>
            <p className="">
                We are always accepting new members into our community. If you would like to join our orchestra or
                get more information about our current performance opportunities, please join us on <a
                href="https://discord.gg/9RVUJMQ" className="text-info">Discord</a>!
            </p>
        </div>
    </div>;
};

const Directors = (props: { directors: Director[] }) => {
    return <div className="row mt-3 border-bottom border-light">
        <div className="col col-lg-3 col-12 text-center">
            <h2>VVGO Leadership</h2>
        </div>
        <div className="col col-lg-9 col-12 text-center mt-2">
            <ExecutiveDirectorTable directors={props.directors}/>
        </div>
    </div>;
};

const ExecutiveDirectorTable = (props: { directors: Director[] }) => {
    return <table id="leader-table" className="table table-responsive table-borderless text-light fs-6">
        <tbody>
        {props.directors.map(director => <ExecutiveDirectorRow key={director.Name} director={director}/>)}
        </tbody>
    </table>;
};

const ExecutiveDirectorRow = (props: { director: Director }) => {
    return <tr className="border-bottom">
        <td><img src={props.director.Icon} alt={props.director.Name} height="100"/></td>
        <td><p className="text-light">{props.director.Name}<br/><small>{props.director.Epithet}</small>
        </p></td>
        <td><p>{props.director.Blurb}</p>
            <p><i>{props.director.Affiliations}</i></p></td>
    </tr>;
};

