// use crypto::digest::Digest;
use crypto::hkdf;
use crypto::hmac::Hmac;
use crypto::mac::Mac;
use crypto::sha2::Sha256;

use util;

const HASH_OUTPUT_SIZE: usize = 32;

pub fn derive_secrets(input: &[u8], info: &[u8]) -> Vec<u8> {
    derive_secrets_with_salt(
        input,
        &Vec::with_capacity(HASH_OUTPUT_SIZE),
        info,
    )
}

pub fn derive_secrets_with_salt(
    input: &[u8],
    salt: &[u8],
    info: &[u8],
) -> Vec<u8> {
    let sha = Sha256::new();
    let mut mac = Hmac::new(sha, salt);
    mac.input(input);

    let mut prk = util::zeroed(32);
    let mut okm = util::zeroed(64);

    hkdf::hkdf_extract(sha, salt, input, &mut prk);
    hkdf::hkdf_expand(sha, &prk, info, &mut okm);

    okm
}
