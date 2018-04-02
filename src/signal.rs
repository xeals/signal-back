#[derive(Clone, PartialEq, Message)]
pub struct SqlStatement {
    #[prost(string, optional, tag = "1")]
    pub statement: ::std::option::Option<String>,
    #[prost(message, repeated, tag = "2")]
    pub parameters: ::std::vec::Vec<sql_statement::SqlParameter>,
}
pub mod sql_statement {
    #[derive(Clone, PartialEq, Message)]
    pub struct SqlParameter {
        #[prost(string, optional, tag = "1")]
        pub string_paramter: ::std::option::Option<String>,
        #[prost(uint64, optional, tag = "2")]
        pub integer_parameter: ::std::option::Option<u64>,
        #[prost(double, optional, tag = "3")]
        pub double_parameter: ::std::option::Option<f64>,
        #[prost(bytes, optional, tag = "4")]
        pub blob_parameter: ::std::option::Option<Vec<u8>>,
        #[prost(bool, optional, tag = "5")]
        pub nullparameter: ::std::option::Option<bool>,
    }
}
#[derive(Clone, PartialEq, Message)]
pub struct SharedPreference {
    #[prost(string, optional, tag = "1")]
    pub file: ::std::option::Option<String>,
    #[prost(string, optional, tag = "2")]
    pub key: ::std::option::Option<String>,
    #[prost(string, optional, tag = "3")]
    pub value: ::std::option::Option<String>,
}
#[derive(Clone, PartialEq, Message)]
pub struct Attachment {
    #[prost(uint64, optional, tag = "1")]
    pub row_id: ::std::option::Option<u64>,
    #[prost(uint64, optional, tag = "2")]
    pub attachment_id: ::std::option::Option<u64>,
    #[prost(uint32, optional, tag = "3")]
    pub length: ::std::option::Option<u32>,
}
#[derive(Clone, PartialEq, Message)]
pub struct Avatar {
    #[prost(string, optional, tag = "1")]
    pub name: ::std::option::Option<String>,
    #[prost(uint32, optional, tag = "2")]
    pub length: ::std::option::Option<u32>,
}
#[derive(Clone, PartialEq, Message)]
pub struct DatabaseVersion {
    #[prost(uint32, optional, tag = "1")]
    pub version: ::std::option::Option<u32>,
}
#[derive(Clone, PartialEq, Message)]
pub struct Header {
    #[prost(bytes, optional, tag = "1")]
    pub iv: ::std::option::Option<Vec<u8>>,
    #[prost(bytes, optional, tag = "2")]
    pub salt: ::std::option::Option<Vec<u8>>,
}
#[derive(Clone, PartialEq, Message)]
pub struct BackupFrame {
    #[prost(message, optional, tag = "1")]
    pub header: ::std::option::Option<Header>,
    #[prost(message, optional, tag = "2")]
    pub statement: ::std::option::Option<SqlStatement>,
    #[prost(message, optional, tag = "3")]
    pub preference: ::std::option::Option<SharedPreference>,
    #[prost(message, optional, tag = "4")]
    pub attachment: ::std::option::Option<Attachment>,
    #[prost(message, optional, tag = "5")]
    pub version: ::std::option::Option<DatabaseVersion>,
    #[prost(bool, optional, tag = "6")]
    pub end: ::std::option::Option<bool>,
    #[prost(message, optional, tag = "7")]
    pub avatar: ::std::option::Option<Avatar>,
}
