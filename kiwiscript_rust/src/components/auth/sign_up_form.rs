use leptos::*;
use serde::{Deserialize, Serialize};

use crate::utils::{get_http_client, FormField};

enum FormFields {
    Email,
    FirstName,
    LastName,
    Location,
    Password1,
    Password2,
}

#[derive(Serialize, Deserialize, Clone)]
pub enum Location {
    #[serde(rename = "NZL")]
    NewZealand,
    #[serde(rename = "AUS")]
    Australia,
    #[serde(rename = "NAM")]
    NorthAmerica,
    #[serde(rename = "EUR")]
    Europe,
    #[serde(rename = "OTH")]
    Other,
}

impl Location {
    pub fn from_str(country: &str) -> Self {
        match country {
            "NZL" => Self::NewZealand,
            "AUS" => Self::Australia,
            "NAM" => Self::NorthAmerica,
            "EUR" => Self::Europe,
            _ => Self::Other,
        }
    }
}

impl ToString for Location {
    fn to_string(&self) -> String {
        match self {
            Self::NewZealand => "NZL".to_string(),
            Self::Australia => "AUS".to_string(),
            Self::NorthAmerica => "NAM".to_string(),
            Self::Europe => "EUR".to_string(),
            Self::Other => "OTH".to_string(),
        }
    }
}

#[derive(Serialize, Deserialize, Clone)]
struct SignUpBody {
    email: String,
    #[serde(rename = "firstName")]
    first_name: String,
    #[serde(rename = "lastName")]
    last_name: String,
    location: Location,
    #[serde(rename = "password")]
    password1: String,
    password2: String,
}

impl SignUpBody {
    fn new() -> Self {
        Self {
            email: String::new(),
            first_name: String::new(),
            last_name: String::new(),
            location: Location::NewZealand,
            password1: String::new(),
            password2: String::new(),
        }
    }
}

#[derive(Clone)]
struct SignUpErrors {
    general: Option<String>,
    email: Option<String>,
    first_name: Option<String>,
    last_name: Option<String>,
    location: Option<String>,
    password1: Option<String>,
    password2: Option<String>,
}

impl SignUpErrors {
    fn new() -> Self {
        Self {
            general: None,
            email: None,
            first_name: None,
            last_name: None,
            location: None,
            password1: None,
            password2: None,
        }
    }
}

async fn submit_form(body: SignUpBody) -> Result<(), reqwest::Error> {
    let url = "http://localhost:5000/api/auth/register";
    get_http_client().post(url).json(&body).send().await?;
    Ok(())
}


fn validate_form(body: &SignUpBody) -> Result<(), String> {
    if body.password1 != body.password2 {
        return Err("Passwords do not match".to_string());
    }

    Ok(())
}

#[component]
pub fn SignUpForm() -> impl IntoView {
    let (email, set_email) = create_signal(FormField::new(String::new()));
    let (first_name, set_first_name) = create_signal(FormField::new(String::new()));
    let (last_name, set_last_name) = create_signal(FormField::new(String::new()));
    let (location, set_location) = create_signal(FormField::new(Location::NewZealand));
    let (password1, set_password1) = create_signal(FormField::new(String::new()));
    let (password2, set_password2) = create_signal(FormField::new(String::new()));

    let (body, set_body) = create_signal(SignUpBody::new());
    let (errors, set_errors) = create_signal(SignUpErrors::new());
    let (loading, set_loading) = create_signal(false);

    let update_body = |field: FormFields| {
        move |e| {
            let val = event_target_value(&e);
            set_body.update(|b| match field {
                FormFields::Email => b.email = val,
                FormFields::FirstName => b.first_name = val,
                FormFields::LastName => b.last_name = val,
                FormFields::Location => b.location = Location::from_str(&val),
                FormFields::Password1 => b.password1 = val,
                FormFields::Password2 => b.password2 = val,
            });
        }
    };


    let submit = create_action(|b: &SignUpBody| {
        let input = b.clone();
        async move {
            let result = submit_form(input).await;

            if let Ok(s) = result {
                println!("Form submitted successfully: {:?}", s);
                tracing::info!("Form submitted successfully: {:?}", s);
            }

            if let Err(e) = result {
                tracing::error!("Error submitting form: {:?}", e);
            }
        }
    });

    let on_submit = move |e: ev::SubmitEvent| {
        e.prevent_default();
        submit.dispatch(body.get());
    };

    view! {
        <form class="container" on:submit=on_submit>
            <div class="row mb-3">
                <label
                    for="email"
                    class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2"
                >
                    Email address
                </label>
                <input
                    on:input=move |e: ev::Event| {
                        let val = event_target_value(&e);
                        set_email.update(|f| f.update(val));
                    }
                    type="email"
                    class="form-control col"
                    id="email"
                    aria-describedby="emailHelp"
                    placeholder="Enter email"
                    prop:value=move || email.get().val()
                />
            </div>
            <div class="row mb-3">
                <label
                    for="first_name"
                    class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2"
                >
                    First name
                </label>
                <input
                    on:input=update_body(FormFields::FirstName)
                    type="text"
                    class="form-control col"
                    id="first_name"
                    placeholder="Enter first name"
                    prop:value=body.with(|b| b.first_name.clone())
                />
            </div>
            <div class="row mb-3">
                <label
                    for="last_name"
                    class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2"
                >
                    Last name
                </label>
                <input
                    on:input=update_body(FormFields::LastName)
                    type="text"
                    class="form-control col"
                    id="last_name"
                    placeholder="Enter last name"
                    prop:value=body.with(|b| b.last_name.clone())
                />
            </div>
            <div class="row mb-3">
                <label
                    for="location"
                    class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2"
                >
                    Location
                </label>
                <select
                    on:input=update_body(FormFields::Location)
                    class="form-select col"
                    id="location"
                    prop:value=body.with(|b| b.location.to_string())
                >
                    <option value="NZL">New Zealand</option>
                    <option value="AUS">Australia</option>
                    <option value="NAM">North America</option>
                    <option value="EUR">Europe</option>
                    <option value="OTH">Other</option>
                </select>
            </div>
            <div class="row mb-3">
                <label
                    for="password1"
                    class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2"
                >
                    Password
                </label>
                <input
                    on:input=update_body(FormFields::Password1)
                    type="password"
                    class="form-control col"
                    id="password1"
                    placeholder="Enter password"
                    prop:value=body.with(|b| b.password1.clone())
                />
            </div>
            <div class="row mb-3">
                <label
                    for="password2"
                    class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2"
                >
                    Confirm Password
                </label>
                <input
                    on:input=update_body(FormFields::Password2)
                    type="password"
                    class="form-control col"
                    id="password2"
                    placeholder="Confirm Password"
                    prop:value=body.with(|b| b.password2.clone())
                />
            </div>
            <div class="row mb-3">
                <button type="submit" class="btn btn-primary col">
                    Submit
                </button>
            </div>
        </form>
    }
}
