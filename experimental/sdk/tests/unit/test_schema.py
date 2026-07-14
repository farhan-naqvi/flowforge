"""Unit tests for schema inference."""

import pytest
from flowforge.schema.inference import infer_schema, validate_data
from typing import List, Dict, Optional


def test_infer_basic_types():
    """Test schema inference for basic types."""
    assert infer_schema(int) == {"type": "integer"}
    assert infer_schema(str) == {"type": "string"}
    assert infer_schema(float) == {"type": "number"}
    assert infer_schema(bool) == {"type": "boolean"}


def test_infer_container_types():
    """Test schema inference for container types."""
    assert infer_schema(list) == {"type": "array"}
    assert infer_schema(dict) == {"type": "object"}


def test_infer_generic_types():
    """Test schema inference for generic types."""
    list_schema = infer_schema(List[int])
    assert list_schema["type"] == "array"
    assert list_schema["items"]["type"] == "integer"
    
    dict_schema = infer_schema(Dict[str, int])
    assert dict_schema["type"] == "object"


def test_infer_optional():
    """Test schema inference for Optional types."""
    schema = infer_schema(Optional[int])
    assert "integer" in str(schema.get("type", ""))
    assert "null" in str(schema.get("type", ""))


def test_infer_complex_types():
    """Test schema inference for complex types."""
    schema = infer_schema(List[Dict[str, int]])
    assert schema["type"] == "array"
    assert schema["items"]["type"] == "object"


def test_infer_none():
    """Test schema inference for None."""
    schema = infer_schema(None)
    assert schema["type"] == "object"


def test_validate_integer():
    """Test validation of integers."""
    schema = {"type": "integer"}
    
    valid, errors = validate_data(42, schema)
    assert valid
    assert len(errors) == 0
    
    valid, errors = validate_data("not an int", schema)
    assert not valid
    assert len(errors) > 0


def test_validate_string():
    """Test validation of strings."""
    schema = {"type": "string"}
    
    valid, errors = validate_data("hello", schema)
    assert valid
    
    valid, errors = validate_data(123, schema)
    assert not valid


def test_validate_array():
    """Test validation of arrays."""
    schema = {
        "type": "array",
        "items": {"type": "integer"},
    }
    
    valid, errors = validate_data([1, 2, 3], schema)
    assert valid
    
    valid, errors = validate_data([1, "not int", 3], schema)
    assert not valid


def test_validate_object():
    """Test validation of objects."""
    schema = {"type": "object"}
    
    valid, errors = validate_data({"key": "value"}, schema)
    assert valid
    
    valid, errors = validate_data("not an object", schema)
    assert not valid


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
