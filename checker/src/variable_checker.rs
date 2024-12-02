use std::{collections::HashMap, process::exit};

use fancy_regex::Regex;
use lazy_static::lazy_static;
use lexer::data_types::is_data_type;
use shared::{code::{function::Function, header::Header, value_name::value_name::VALUE_NAME_REGEX}, logger::{Logger, LoggerImpl}, result::try_unwrap, token::{token::{Token, TokenImpl}, token_type::TokenType}};

use crate::{header_checker::{check_header_value_definition, find_imported_classes}, scope_checker::throw_value_already_defined};

lazy_static! {
    // Used to print warnings for cammel case variable names
    // Surf encourages snake case variable names!
    pub static ref CAMMEL_CASE_REGEX: Regex = 
        Regex::new(r"^[a-zA-Z][a-zA-Z0-9]*$").unwrap();
}

fn check_variable_name(var_name: &String, trace: &String) {
    if try_unwrap(
        CAMMEL_CASE_REGEX.is_match(var_name),
        "Failed to validate a variable name"
    ) {
        Logger::warn(
            "Consider using snake case for variable names",
            &[
                format!(
                    "Consider converting {} to snake case",
                    var_name
                ).as_str(),
                trace.as_str()
            ],
        );
    }
}

pub fn check_variables(
    tokens: &Vec<Token>,
    start: &usize,
    // Used to check if a value is already defined
    functions: &HashMap<String, Function>,
    imports: &Vec<Header>
) {
    // Variable definitions should be already validated by now
    // Example definition:
    // let my_var : str = "Hello, world!";
    // Number of tokens: 7

    // The first token should be the variable name (We don't receive the let token)
    let variable_tokens = tokens[*start..].to_vec();
    if variable_tokens.len() < 4 {
        Logger::err(
            "Invalid variable definition",
            &[
                "Variable definitions must have at least 4 tokens"
            ],
            &[
                tokens[*start].build_trace().as_str()
            ]
        );

        exit(1);
    }

    let var_name = &variable_tokens[0];
    let var_name_value = var_name.get_value();
    let colon = &variable_tokens[1];
    let var_type = &variable_tokens[2];
    let equals = &variable_tokens[3];

    // The var name should be an unknown token (not a keyword)
    if
        var_name.get_token_type() != TokenType::Unknown
        || !try_unwrap(
            VALUE_NAME_REGEX.is_match(var_name.get_value().as_str()),
            "Failed to validate a variable name"
        )
    {
        Logger::err(
            "Invalid variable name",
            &[
                "Variable names must be unknown tokens"
            ],
            &[
                var_name.build_trace().as_str()
            ]
        );

        exit(1);
    }

    // The colon should be a colon
    if colon.get_token_type() != TokenType::Colon {
        Logger::err(
            "Invalid variable definition",
            &[
                "Expected a colon after the variable name"
            ],
            &[
                colon.build_trace().as_str()
            ]
        );

        exit(1);
    }

    // The equals should be an equals
    if equals.get_token_type() != TokenType::Assign {
        Logger::err(
            "Invalid variable definition",
            &[
                "Expected an equals sign after the variable type"
            ],
            &[
                equals.build_trace().as_str()
            ]
        );

        exit(1);
    }

    // Check if the var name is already defined
    if
        functions.contains_key(var_name_value.as_str()) ||
        check_header_value_definition(&var_name_value, imports)
    {
        throw_value_already_defined(
            &var_name_value,
            &var_name.build_trace()
        );
    }

    // Check if the type is a data type or an imported object
    if !is_data_type(var_type.get_token_type()) {
        // Check if the variable name is in cammel case
        let class_optional = find_imported_classes(
            &var_type.get_value(),
            imports
        );

        if class_optional.is_none() {
            Logger::err(
                "Invalid data type",
                &[
                    "The data type is not recognized"
                ],
                &[
                    var_type.build_trace().as_str()
                ]
            );

            exit(1);
        }

        return;
    }
        
}