import React from "react";
import {Link} from "react-router-dom";
import styles from '../css/error_page.module.css'
import {Helmet} from "react-helmet";
import notFoundGif from '../images/404.gif'
import accessDeniedGif from '../images/401.gif'
import internalOopsieGif from '../images/500.gif'

function NotFound() {
    return <div>
        <Helmet>
            <title>404 Not Found</title>
            <meta name="description" content=""/>
        </Helmet>
        <img className={styles.errorImg} src={notFoundGif} alt="404 Not Found"/>
        <div className={styles.helpMe}>
            <Link to="/" className={styles.helpMe}>Click here to return to safety.</Link>
        </div>
    </div>
}

function AccessDenied() {
    return <div>
        <Helmet>
            <title>401 Access Denied</title>
            <meta name="description" content=""/>
        </Helmet>
        <img className={styles.errorImg} src={accessDeniedGif} alt="404 Not Found"/>
        <div className={styles.helpMe}>
            <Link to="/" className={styles.helpMe}>Click here to return to safety.</Link>
        </div>
    </div>
}

function InternalOopsie() {
    return <div>
        <Helmet>
            <title>500 Internal Oopsie</title>
            <meta name="description" content=""/>
        </Helmet>
        <img className={styles.errorImg} src={internalOopsieGif} alt="404 Not Found"/>
        <div className={styles.helpMe}>
            <Link to="/" className={styles.helpMe}>Click here to return to safety.</Link>
        </div>
    </div>
}

export {NotFound, AccessDenied, InternalOopsie}
