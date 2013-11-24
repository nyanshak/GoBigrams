Go Bigram Model
===============

This is a very simplistic bigram training and testing program. It was made for a small NLP class project.

*bigramTrain* outputs a language model, given a training text.
*bigramTest* gives the perplexity of an input file, given a language model.

Usage for *bigramTrain*:
	bigramTrain -lm <lm output file> -text <training input text>
Usage for *bigramText*:
	bigramTest -lm <lm input file> -text <test input text>
