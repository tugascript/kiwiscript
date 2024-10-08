use leptos::*;

enum FormFields {
    Email,
    FirstName,
    LastName,
    Location,
    Password1,
    Password2,
}

pub enum Location {
    NewZealand,
    Australia,
    NorthAmerica,
    Europe,
    Other,
}

impl Location {
    pub fn to_string(&self) -> String {
        match self {
            Self::NewZealand => "NZL".to_string(),
            Self::Australia => "AUS".to_string(),
            Self::NorthAmerica => "NAM".to_string(),
            Self::Europe => "EUR".to_string(),
            Self::Other => "OTH".to_string(),
        }
    }

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

struct SignUpBody {
    email: String,
    first_name: String,
    last_name: String,
    location: Location,
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

#[component]
pub fn SignUpForm() -> impl IntoView {
    let (body, set_body) = create_signal(SignUpBody::new());
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

    view! {
        <form class="container">
            <div class="row mb-3">
                <label for="email" class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2">
                    Email address
                </label>
                <input
                    on:input=update_body(FormFields::Email)
                    type="email"
                    class="form-control col"
                    id="email"
                    aria-describedby="emailHelp"
                    placeholder="Enter email"
                    prop:value=body.with(|b| b.email.clone())
                />
            </div>
            <div class="row mb-3">
                <label for="first_name" class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2">
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
                <label for="last_name" class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2">
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
                <label for="location" class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2">
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
                <label for="password1" class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2">
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
                <label for="password2" class="form-label col-xs-12 col-sm-12 col-md-3 col-lg-2 col-xl-2">
                    Confirm Password
                </label>
                <input
                    on:input=update_body(FormFields::Password2)
                    type="password"
                    class="form-control col"
                    id="password1"
                    placeholder="Confirm Password"
                    prop:value=body.with(|b| b.password2.clone())
                />
            </div>
        </form>
    }
}
